// Nuvoton NUC980 ETimer support
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
//   Timer0 eclk = XIN (set in EarlyClockInit)
//   PRESCALE = 11 → tick rate = 12,000,000 / (11+1) = 1,000,000 Hz = 1 MHz
//   Each count = 1 µs; multiply by 1000 for nanoseconds.

package nuc980

import (
	"github.com/usbarmory/tamago/soc/nuvoton/etimer"
)

// ETimer register bases (ETimer1 stride = 0x100).
const (
	ETMR0_BA = 0xb0050000
	ETMR1_BA = ETMR0_BA + 0x100
)

// timerPrescale yields a 1 MHz tick from the 12 MHz XIN eclk: 12 / (11+1).
const timerPrescale = 11

// ETimer1 periodic interrupt interval in microseconds (10 ms at 1 MHz).
const ETMR1_PERIOD_US = 10000

// ETimer0 is the free-running counter backing nanotime; ETimer1 is the
// periodic scheduler interrupt source.
var (
	ETimer0 = &etimer.ETimer{Base: ETMR0_BA}
	ETimer1 = &etimer.ETimer{Base: ETMR1_BA}
)

// timerWrapPeriod is the number of ticks in one full counter cycle.
const timerWrapPeriod = uint64(etimer.CMPR_MAX) + 1

var (
	// timerLast holds the most recent hardware counter value seen by
	// readTimerExtended.
	timerLast uint32
	// timerHigh accumulates full counter periods (each 2^24 ticks).
	timerHigh uint64
	// timerStarted guards initTimer against re-running, which would reset
	// the free-running counter (and nanotime) mid-boot.
	timerStarted bool
)

// readTimerExtended returns a 64-bit microsecond count by extending the
// 24-bit hardware counter with software wrap-around detection. It must be
// called at least once per counter period (~16.77 s); the runtime scheduler
// satisfies this by calling nanotime() continuously.
func readTimerExtended() uint64 {
	now := ETimer0.Count()

	if now < timerLast {
		timerHigh += timerWrapPeriod
	}

	timerLast = now
	return timerHigh + uint64(now)
}

// initTimer configures ETimer0 as a 1 MHz free-running up-counter.
//
// After this call, readTimerExtended() returns microseconds elapsed since
// init. The APB clock for Timer0 and the XIN eclk mux are configured in
// EarlyClockInit.
func initTimer() {
	if timerStarted {
		return
	}
	timerStarted = true

	ETimer0.Stop()
	ETimer0.SetPrescale(timerPrescale)
	ETimer0.SetCompare(etimer.CMPR_MAX)
	ETimer0.EnableInterrupt(false)
	ETimer0.ClearInterrupt()

	// Reset wrap-tracking state.
	timerLast = 0
	timerHigh = 0

	// Enable in periodic mode with DBGACK. With CMPR at maximum
	// (0xffffff), periodic mode wraps every ~16.8 s — effectively
	// free-running. Continuous mode causes a boot hang on NUC980 silicon
	// for reasons under investigation.
	ETimer0.Start(etimer.CTL_DBGACK | etimer.CTL_PERIODIC)
}

// InitInterruptTimer configures ETimer1 as a periodic interrupt source at the
// specified period (in microseconds). On each compare match the counter
// resets and raises a periodic IRQ through the AIC (IRQ_ETMR1).
//
// ETimer1's APB clock and eclk source are configured from assembly in the
// board cpuinit via EarlyClockInit. The timer is not started here; call
// StartInterruptTimer once ServiceInterrupts is running.
func InitInterruptTimer(periodUs uint32) {
	ETimer1.Stop()
	ETimer1.SetPrescale(timerPrescale)
	ETimer1.SetCompare(periodUs - 1)
	ETimer1.ClearInterrupt()
	ETimer1.EnableInterrupt(true)
}

// StartInterruptTimer enables ETimer1 in periodic mode. Call this only after
// arm.ServiceInterrupts is running and the signal relay path is fully
// initialized; starting the timer earlier causes a boot hang because the
// periodic interrupt fires before the relay goroutine is ready.
func StartInterruptTimer() {
	ETimer1.ClearInterrupt()
	ETimer1.Start(etimer.CTL_PERIODIC)
}
