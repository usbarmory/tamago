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

const (
	// p178, Table 2-3, IMX6ULLRM
	SYS_CNT_BASE uint32 = 0x021dc000

	// p2402, Table D5-1, ARMv7 Architecture Reference Manual
	CNTCR   = SYS_CNT_BASE
	CNTSR   = SYS_CNT_BASE + 0x04
	CNTCV1  = SYS_CNT_BASE + 0x08
	CNTCV2  = SYS_CNT_BASE + 0x0c
	CNTFID0 = SYS_CNT_BASE + 0x20 // base frequency
	CNTFID1 = SYS_CNT_BASE + 0x24 // frequency divider 1
	CNTFID2 = SYS_CNT_BASE + 0x28 // frequency divider 2
	CNTID   = SYS_CNT_BASE + 0xfd0

	// p2410, D5.7.2 CNTCR, Counter Control Register, ARMv7 Architecture
	// Reference Manual
	CNTCR_FCREQ2 = 10 // frequency = CNTFID0/CNTFID2
	CNTCR_FCREQ1 = 9  // frequency = CNTFID0/CNTFID1
	CNTCR_FCREQ0 = 8  // frequency = CNTFID0
	CNTCR_HDBG   = 1
	CNTCR_EN     = 0

	// nanoseconds
	refFreq int64 = 1000000000
)

// defined in timer_arm.s
func read_gtc() int64
func read_cntfrq() int32
func write_cntfrq(int32)
func read_cntpct() int64
func Busyloop(int32)

// InitGlobalTimers initializes ARM Cortex-A9 timers
func (c *CPU) InitGlobalTimers() {
	c.TimerFn = read_gtc
	c.TimerMultiplier = 10
}

// InitGenericTimers initializes ARM Cortex-A7 timers
func (c *CPU) InitGenericTimers(freq int32) {
	var timerFreq int64

	if freq != 0 {
		write_cntfrq(freq)
		// Set base frequency
		reg.Write(CNTFID0, uint32(freq))
		// Set system counter to base frequency
		reg.Set(CNTCR, CNTCR_FCREQ0)
		// Stop system counter on debug
		reg.Set(CNTCR, CNTCR_HDBG)
		// Start system counter
		reg.Set(CNTCR, CNTCR_EN)
	}

	timerFreq = int64(read_cntfrq())

	c.TimerMultiplier = int64(refFreq / timerFreq)
	c.TimerFn = read_cntpct
}
