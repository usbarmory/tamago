// microvm support for tamago/amd64
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package microvm

import (
	"encoding/binary"
	"math/big"
	"time"
	_ "unsafe"

	"github.com/usbarmory/tamago/dma"
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

//go:linkname nanotime1 runtime.nanotime1
func nanotime1() int64 {
	return int64(float64(AMD64.TimerFn())*AMD64.TimerMultiplier) + AMD64.TimerOffset
}

// defined in timer.s
func pvclock(msr uint32) (ptr uint32)

var (
	// TimeInfoUpdate is the kvmclock sync interval
	TimeInfoUpdate time.Duration = 1 * time.Second
	timeInfoBuffer []byte
	timeInfo       *pvClockTimeInfo
)

func initTimeInfo(msr uint32) {
	ptr := pvclock(msr)
	size := 32

	r, err := dma.NewRegion(uint(ptr), size, true)

	if err != nil {
		panic("internal error")
	}

	_, timeInfoBuffer = r.Reserve(size, 0)
	timeInfo = &pvClockTimeInfo{}
}

func kvmClock() int64 {
	binary.Decode(timeInfoBuffer, binary.LittleEndian, timeInfo)

	delta := AMD64.TimerFn() - timeInfo.Timestamp

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

// As nanotime1() cannot malloc the sync needs to update asynchronously (or be
// moved to Go assembly).
func kvmClockSync() {
	var version uint32

	for {
		time.Sleep(TimeInfoUpdate)

		binary.Decode(timeInfoBuffer, binary.LittleEndian, timeInfo)

		if timeInfo.Version == version || timeInfo.Version%2 == 1 {
			continue
		}

		version = timeInfo.Version
		AMD64.SetTimer(kvmClock())
	}
}

func init() {
	features := AMD64.Features()

	switch {
	case features.InvariantTSC && !features.KVM:
		// no action required
	case features.InvariantTSC && features.KVM:
		initTimeInfo(features.KVMClockMSR)
		// sync to kvmclock once
		AMD64.SetTimer(kvmClock())
	case features.KVM && features.KVMClockMSR > 0:
		initTimeInfo(features.KVMClockMSR)
		// sync to kvmclock every TimeInfoUpdate
		go kvmClockSync()
	default:
		panic("could not set system timer")
	}
}
