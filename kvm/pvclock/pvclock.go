// KVM pvclock driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package pvclock implements a driver for the KVM specific paravirtualized
// clocksource following the MSR_KVM_SYSTEM_TIME_NEW and MSR_KVM_WALL_CLOCK_NEW
// KVM-specific MSRs as described at:
//
//	https://docs.kernel.org/virt/kvm/x86/msr.html
//
// This package is only meant to be used with `GOOS=tamago GOARCH=amd64` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package pvclock

import (
	"encoding/binary"
	"math/big"
	"sync"
	"time"

	"github.com/usbarmory/tamago/amd64"
	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/internal/reg"
)

type pvClockTimeInfo struct {
	Version    uint32
	_          uint32
	Timestamp  uint64
	SystemTime uint64
	Multiplier uint32
	Shift      int8
	Flags      uint8
	_          [2]uint8
}

type pvClockWallClock struct {
	Version uint32
	Sec     uint32
	Nsec    uint32
}

func pvclock(msr uint32) (ptr uint32)

var (
	// TimeInfoUpdate is the kvmclock sync interval
	TimeInfoUpdate time.Duration = 1 * time.Second

	// host shared DMA buffers
	timeInfoBuffer  []byte
	wallClockAddr   uint
	wallClockBuffer []byte

	wallClockCPU *amd64.CPU
	wallClockMSR uint32
	wallClockMu  sync.Mutex
)

func initTimeInfo(msr uint32) {
	ptr := pvclock(msr)
	size := 32

	r, err := dma.NewRegion(uint(ptr), size, true)

	if err != nil {
		panic("internal error")
	}

	_, timeInfoBuffer = r.Reserve(size, 0)
}

func initWallClock(cpu *amd64.CPU, msr uint32) {
	wallClockCPU = cpu
	wallClockMSR = msr
	wallClockAddr, wallClockBuffer = dma.Reserve(16, 4)
}

// wallClock returns the host wall clock time in nanoseconds, KVM computes it
// when the MSR is written.
func wallClock() int64 {
	wallClockMu.Lock()
	defer wallClockMu.Unlock()

	reg.WriteMSR(uint64(wallClockMSR), uint64(wallClockAddr))

	wc := &pvClockWallClock{}
	binary.Decode(wallClockBuffer, binary.LittleEndian, wc)

	return int64(wc.Sec)*int64(time.Second) + int64(wc.Nsec) + pvClock(wallClockCPU, nil)
}

// Now returns the KVM host wall clock time, or the zero time if Init found no
// kvmclock.
func Now() time.Time {
	if wallClockBuffer == nil {
		return time.Time{}
	}

	return time.Unix(0, wallClock())
}

func pvClock(cpu *amd64.CPU, timeInfo *pvClockTimeInfo) int64 {
	if timeInfo == nil {
		timeInfo = &pvClockTimeInfo{}
	}

	binary.Decode(timeInfoBuffer, binary.LittleEndian, timeInfo)
	delta := cpu.Counter() - timeInfo.Timestamp

	if timeInfo.Shift < 0 {
		delta >>= -timeInfo.Shift
	} else {
		delta <<= timeInfo.Shift
	}

	d := big.NewInt(int64(delta))
	m := big.NewInt(int64(timeInfo.Multiplier))
	r := big.NewInt(0)

	r.Mul(d, m)
	r.Rsh(r, 32)

	return int64(r.Uint64() + timeInfo.SystemTime)
}

// pvClockSync adjusts the CPU system timer offset by the drift of the TSC
// against the KVM host wall clock, so that a time set by the application is
// kept.
func pvClockSync(cpu *amd64.CPU) {
	version := uint32(0)
	timeInfo := &pvClockTimeInfo{}

	drift := func() int64 {
		return wallClock() - int64(float64(cpu.Counter())*cpu.TimerMultiplier)
	}

	last := drift()

	for {
		time.Sleep(TimeInfoUpdate)

		binary.Decode(timeInfoBuffer, binary.LittleEndian, timeInfo)

		if timeInfo.Version == version || timeInfo.Version%2 == 1 {
			continue
		}

		version = timeInfo.Version

		d := drift()
		cpu.TimerOffset += d - last
		last = d
	}
}

// Init adjusts the CPU system timer using the KVM pvclock as required by the
// Time Stamp Counter (TSC) reliability.
func Init(cpu *amd64.CPU) {
	features := cpu.Features()

	switch {
	case features.TSCInvariant && !features.KVM:
		// no action required as TSC is reliable
	case features.TSCInvariant && features.KVM && features.KVMSystemTimeMSR > 0:
		// no action required as TSC is reliable but we
		// opportunistically adjust once with kvmclock
		initTimeInfo(features.KVMSystemTimeMSR)
		initWallClock(cpu, features.KVMWallClockMSR)
		cpu.SetTime(pvClock(cpu, nil))
	case features.KVM && features.KVMSystemTimeMSR > 0:
		// TSC must be adjusted as it is not reliable through state
		// changes.
		//
		// As nanotime() cannot malloc we cannot override it, rather
		// we adjust asynchronously with kvmclock every TimeInfoUpdate
		// interval.
		//
		// If ever required pvClockSync() can be moved to Go assembly.
		initTimeInfo(features.KVMSystemTimeMSR)
		initWallClock(cpu, features.KVMWallClockMSR)
		cpu.SetTime(pvClock(cpu, nil))
		go pvClockSync(cpu)
	default:
		panic("could not set system timer")
	}
}
