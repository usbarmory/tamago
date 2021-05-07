// ARM processor support
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm

import (
	_ "unsafe"

	"github.com/f-secure-foundry/tamago/internal/reg"
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
	refFreq int64 = 1000000000
)

// defined in timer_arm.s
func read_gtc() int64
func read_cntfrq() int32
func write_cntfrq(freq int32)
func write_cntkctl(val uint32)
func read_cntpct() int64

// Busyloop spins the processor for busy waiting purposes, taking a counter
// value for the number of loops.
func Busyloop(count int32)

// InitGlobalTimers initializes ARM Cortex-A9 timers.
func (cpu *CPU) InitGlobalTimers() {
	cpu.TimerFn = read_gtc
	cpu.TimerMultiplier = 10
}

// InitGenericTimers initializes ARM Cortex-A7 timers.
func (cpu *CPU) InitGenericTimers(base uint32, freq int32) {
	var timerFreq int64

	if freq != 0 {
		// set base frequency
		write_cntfrq(freq)
		reg.Write(base+CNTFID0, uint32(freq))

		// set system counter to base frequency
		reg.Set(base+CNTCR, CNTCR_FCREQ0)
		// stop system counter on debug
		reg.Set(base+CNTCR, CNTCR_HDBG)
		// start system counter
		reg.Set(base+CNTCR, CNTCR_EN)

		// grant PL0 access
		write_cntkctl(1 << CNTKCTL_PL0PCTEN)
	}

	timerFreq = int64(read_cntfrq())

	cpu.TimerMultiplier = int64(refFreq / timerFreq)
	cpu.TimerFn = read_cntpct
}
