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

// defined in timer.s
func pvclock(msr uint32) (ptr uint32)

var (
	// TimeInfoUpdate is the kvmclock sync interval
	TimeInfoUpdate time.Duration = 1 * time.Second

	// host shared DMA buffer
	timeInfoBuffer []byte
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

func kvmClock(timeInfo *pvClockTimeInfo) int64 {
	if timeInfo == nil {
		timeInfo = &pvClockTimeInfo{}
	}

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

func kvmClockSync() {
	version := uint32(0)
	timeInfo := &pvClockTimeInfo{}

	for {
		time.Sleep(TimeInfoUpdate)

		binary.Decode(timeInfoBuffer, binary.LittleEndian, timeInfo)

		if timeInfo.Version == version || timeInfo.Version%2 == 1 {
			continue
		}

		version = timeInfo.Version
		AMD64.SetTimer(kvmClock(timeInfo))
	}
}

func init() {
	features := AMD64.Features()

	switch {
	case features.InvariantTSC && !features.KVM:
		// no action required as TSC is reliable
	case features.InvariantTSC && features.KVM && features.KVMClockMSR > 0:
		// no action required as TSC is reliable but we
		// opportunistically adjust once with kvmclock
		initTimeInfo(features.KVMClockMSR)
		AMD64.SetTimer(kvmClock(nil))
	case features.KVM && features.KVMClockMSR > 0:
		// TSC must be adjusted as it is not reliable through state
		// changes.
		//
		// As nanotime1() cannot malloc we cannot override it rather we
		// adjust asynchronously with kvmclock every TimeInfoUpdate
		// interval. If ever required kvmClockSync() can be moved to Go
		// assembly.
		initTimeInfo(features.KVMClockMSR)
		go kvmClockSync()
	default:
		panic("could not set system timer")
	}
}
