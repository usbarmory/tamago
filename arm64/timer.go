// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm64

import (
	"math"

	"github.com/usbarmory/tamago/internal/reg"
)

// ARM timer register constants
// (ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile)
const (
	// p6721, Table 12-2
	CNTCR = 0
	// base frequency
	CNTFID0 = 0x20

	// p6855, I5.7.2 CNTCR, Counter Control Register
	CNTCR_FCREQ = 8
	CNTCR_HDBG  = 1
	CNTCR_EN    = 0

	CNTKCTL_PL0PCTEN = 0

	// nanoseconds
	refFreq int64 = 1e9
)

// Interrupts
const TIMER_IRQ = 30

// defined in timer.s
func read_cntfrq() uint32
func write_cntfrq(freq uint32)
func write_cntkctl(val uint32)
func read_cntpct() uint64
func write_cntptval(val uint32, enable bool)

// InitGenericTimers initializes ARMv8 Generic Timers.
func (cpu *CPU) InitGenericTimers(base uint32, freq uint32) {
	if freq != 0 {
		// set base frequency
		write_cntfrq(freq)

		if base != 0 {
			reg.Write(base+CNTFID0, freq)

			// set system counter to base frequency
			reg.Set(base+CNTCR, CNTCR_FCREQ)
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

// SetAlarm sets a physical timer to the absolute time matching the argument
// nanoseconds value, an interrupt is generated at expiration.
func (cpu *CPU) SetAlarm(ns int64) {
	if ns == 0 {
		write_cntptval(0, false)
		return
	}

	if cpu.TimerMultiplier == 0 {
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
