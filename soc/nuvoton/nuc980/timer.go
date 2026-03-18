// Nuvoton NUC980 ETimer driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// ETimer0 free-running counter for nanotime() on the NUC980 SoC.
//
// Clock chain:
//   XIN crystal = 12,000,000 Hz
//   CLK_DIVCTL8[17:16] = 0b00 → Timer0 eclk = XIN (set by Init())
//   PRESCALE = 11 → tick rate = 12,000,000 / (11+1) = 1,000,000 Hz = 1 MHz
//   Each DR count = 1 µs; multiply by 1000 for nanoseconds.
//
// Register references:
//   NUC980 Series Datasheet, p. 173 (§ 6.3.6, CLK_DIVCTL8 register)
//   NUC980 Series Datasheet, p. 185 (§ 6.9 Timer Controller)

package nuc980

import (
	"github.com/usbarmory/tamago/internal/reg"
)

// ETimer0 register addresses
//
// NUC980 Series Datasheet, p. 185 (§ 6.9 Timer Controller).
const (
	ETMR0_BA = 0xB0050000

	ETMR_CTL    = ETMR0_BA + 0x00 // Timer Control Register
	ETMR_PRECNT = ETMR0_BA + 0x04 // Pre-scale Counter Register
	ETMR_CMPR   = ETMR0_BA + 0x08 // Compare Register
	ETMR_IER    = ETMR0_BA + 0x0C // Interrupt Enable Register
	ETMR_ISR    = ETMR0_BA + 0x10 // Interrupt Status Register
	ETMR_DR     = ETMR0_BA + 0x14 // Data Register (current count)
)

// CTL register bits
const (
	ETMR_CTL_EN   = 1 << 0   // Timer enable
	ETMR_CTL_CONT = 0x3 << 3 // Mode bits [4:3] = 0b11 (continuous/free-running)
)

// CMPR maximum value: 24-bit compare register.
const ETMR_CMPR_MAX = 0x00FFFFFF

// timerWrapPeriod is the number of ticks in one full counter cycle.
const timerWrapPeriod = uint64(ETMR_CMPR_MAX) + 1

// timerLast holds the most recent hardware counter value seen by
// readTimerExtended.
var timerLast uint32

// timerHigh accumulates full counter periods (each 2^24 ticks = ~16.77 s).
var timerHigh uint64

// readTimerExtended returns a 64-bit microsecond count by extending the
// 24-bit hardware counter with software wrap-around detection.  It must
// be called at least once per counter period (~16.77 s); the runtime
// scheduler satisfies this by calling nanotime() continuously.
func readTimerExtended() uint64 {
	// Mask to 24 bits: the datasheet defines TIMERx_CNT[23:0] as the
	// counter field; bits [31:24] are reserved.
	now := readTimer() & 0x00FFFFFF

	if now < timerLast {
		timerHigh += timerWrapPeriod
	}

	timerLast = now
	return timerHigh + uint64(now)
}

// initTimer configures ETimer0 as a 1 MHz free-running up-counter.
//
// After this call, readTimerExtended() returns microseconds elapsed
// since init.  The APB clock for Timer0 and the XIN eclk mux are
// configured in Init().
func initTimer() {
	// Disable timer before reconfiguring.
	reg.Write(ETMR_CTL, 0)

	// Prescale = 11: tick rate = XIN / (11+1) = 12 MHz / 12 = 1 MHz.
	reg.Write(ETMR_PRECNT, 11)

	// Set compare to maximum (24-bit) so the counter wraps at 16.77 s.
	reg.Write(ETMR_CMPR, ETMR_CMPR_MAX)

	// Disable interrupts.
	reg.Write(ETMR_IER, 0)
	reg.Write(ETMR_ISR, 0x1) // clear any pending

	// Reset wrap-tracking state.
	timerLast = 0
	timerHigh = 0

	// Enable in continuous (free-running) mode: CTL = EN | CONT.
	reg.Write(ETMR_CTL, ETMR_CTL_EN|ETMR_CTL_CONT)
}

// defined in timer.s
func readTimer() uint32
