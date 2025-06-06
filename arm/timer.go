// ARM processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm

import (
	"math"

	"github.com/usbarmory/tamago/internal/reg"
)

// ARM timer register constants
const (
	// p2402, Table D5-1, ARMv7 Architecture Reference Manual
	CNTCR = 0
	// base frequency
	CNTFID0 = 0x20

	// p2410, D5.7.2 CNTCR, Counter Control Register, ARMv7 Architecture
	// Reference Manual
	//
	// frequency = CNTFID0/CNTFID2
	CNTCR_FCREQ2 = 10
	// frequency = CNTFID0/CNTFID1
	CNTCR_FCREQ1 = 9
	// frequency = CNTFID0
	CNTCR_FCREQ0 = 8
	CNTCR_HDBG   = 1
	CNTCR_EN     = 0

	CNTKCTL_PL0PCTEN = 0

	// nanoseconds
	refFreq int64 = 1e9
)

// defined in timer.s
func read_cntfrq() uint32
func write_cntfrq(freq uint32)
func write_cntkctl(val uint32)
func read_cntpct() uint64
func write_cntptval(val uint32, enable bool)

// Busyloop spins the processor for busy waiting purposes, taking a counter
// value for the number of loops.
func Busyloop(count uint32)

// InitGenericTimers initializes ARM Cortex-A7 timers.
func (cpu *CPU) InitGenericTimers(base uint32, freq uint32) {
	if freq != 0 && cpu.Secure() {
		// set base frequency
		write_cntfrq(freq)

		if base != 0 {
			reg.Write(base+CNTFID0, freq)

			// set system counter to base frequency
			reg.Set(base+CNTCR, CNTCR_FCREQ0)
			// stop system counter on debug
			reg.Set(base+CNTCR, CNTCR_HDBG)
			// start system counter
			reg.Set(base+CNTCR, CNTCR_EN)
		}

		// grant PL0 access
		write_cntkctl(1 << CNTKCTL_PL0PCTEN)
	}

	cpu.TimerMultiplier = float64(refFreq) / float64(read_cntfrq())
}

// Counter returns the CPU Counter-timer Physical Count (CNTPCT).
func (cpu *CPU) Counter() uint64 {
	return read_cntpct()
}

// GetTime returns the system time in nanoseconds.
func (cpu *CPU) GetTime() int64 {
	return int64(float64(cpu.Counter())*cpu.TimerMultiplier) + cpu.TimerOffset
}

// SetTime adjusts the system time to the argument nanoseconds value.
func (cpu *CPU) SetTime(ns int64) {
	if cpu.TimerMultiplier == 0 {
		return
	}

	cpu.TimerOffset = ns - int64(float64(read_cntpct())*cpu.TimerMultiplier)
}

// SetAlarm sets a physical timer countdown to the absolute time matching the
// argument nanoseconds value.
func (cpu *CPU) SetAlarm(ns int64) {
	if ns == 0 {
		write_cntptval(0, false)
		return
	}

	set := uint64(ns) / uint64(cpu.TimerMultiplier)
	now := read_cntpct()
	cnt := set - now

	if set <= now {
		cnt = 1
	} else if cnt > math.MaxInt32 {
		cnt = math.MaxInt32
	}

	write_cntptval(uint32(cnt), true)
}
