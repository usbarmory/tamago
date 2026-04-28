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

	ETMR1_BA = ETMR0_BA + 0x100 // ETimer1 base (stride = 0x100)

	ETMR1_CTL    = ETMR1_BA + 0x00 // Timer Control Register
	ETMR1_PRECNT = ETMR1_BA + 0x04 // Pre-scale Counter Register
	ETMR1_CMPR   = ETMR1_BA + 0x08 // Compare Register
	ETMR1_IER    = ETMR1_BA + 0x0C // Interrupt Enable Register
	ETMR1_ISR    = ETMR1_BA + 0x10 // Interrupt Status Register
	ETMR1_DR     = ETMR1_BA + 0x14 // Data Register (current count)
)

// CTL register bits
//
// Linux: drivers/misc/nuc980-etimer.c
const (
	ETMR_CTL_EN         = 0x01 // bit 0: timer enable (CEN)
	ETMR_CTL_DBGACK     = 0x08 // bit 3: continue counting during ICE debug halt
	ETMR_CTL_PERIODIC   = 0x10 // bits [5:4] = 01: periodic mode
	ETMR_CTL_TOGGLE     = 0x20 // bits [5:4] = 10: toggle mode
	ETMR_CTL_CONTINUOUS = 0x30 // bits [5:4] = 11: continuous (free-running)
)

// CMPR maximum value: 24-bit compare register.
const ETMR_CMPR_MAX = 0x00FFFFFF

// ETimer1 periodic interrupt interval in microseconds.
// 10,000 µs = 10 ms at 1 MHz tick rate.
const ETMR1_PERIOD_US = 10000

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

	// Enable timer in periodic mode with DBGACK.  With CMPR at maximum
	// (0xFFFFFF), periodic mode wraps every ~16.8 s — effectively
	// free-running.  Continuous mode (0x31) causes a boot hang on
	// NUC980 silicon for reasons under investigation.
	reg.Write(ETMR_CTL, ETMR_CTL_EN|ETMR_CTL_DBGACK|ETMR_CTL_PERIODIC)
}

// defined in timer.s
func readTimer() uint32

// InitInterruptTimer configures ETimer1 as a periodic interrupt source
// at the specified period (in microseconds).  The timer compare-reset
// mode resets the counter on each compare match, producing a periodic
// IRQ through the AIC (IRQ_ETMR1 = 17).
//
// ETimer1's APB clock and eclk source are configured from assembly
// in the board cpuinit via EarlyInit.
func InitInterruptTimer(periodUs uint32) {
	// Disable timer before reconfiguring.
	reg.Write(ETMR1_CTL, 0)

	// Prescale = 11: tick rate = XIN / (11+1) = 1 MHz.
	reg.Write(ETMR1_PRECNT, 11)

	// Compare value: fire every periodUs ticks at 1 MHz.
	reg.Write(ETMR1_CMPR, periodUs-1)

	// Clear any pending interrupt.
	reg.Write(ETMR1_ISR, 0x1)

	// Enable compare-match interrupt.
	reg.Write(ETMR1_IER, 0x1)

}

// StartInterruptTimer enables ETimer1 in periodic mode.  Call this
// only after arm.ServiceInterrupts is running and the signal relay
// path is fully initialized; starting the timer earlier causes a
// boot hang because the periodic interrupt fires before the relay
// goroutine is ready.
func StartInterruptTimer() {
	// Clear any ISR flag that may have been latched between
	// InitInterruptTimer and now (defensive, not expected).
	reg.Write(ETMR1_ISR, 0x1)
	reg.Write(ETMR1_CTL, ETMR_CTL_EN|ETMR_CTL_PERIODIC)
}
