// ARM Global and Generic timers
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build tamago,arm

package arm

import (
	_ "unsafe"
)

// nanoseconds
const refFreq int64 = 1000000000

var TimerFn func() int64
var timerMultiplier int64

// defined in timer_arm.s
func read_gtc() int64
func read_cntpct() int64
func Busyloop(int32)

// InitGlobalTimers initializes ARM Cortex-A9 timers
func InitGlobalTimers() {
	TimerFn = read_gtc
	timerMultiplier = 10
}

// InitGenericTimers initializes ARM Cortex-A7 timers
func InitGenericTimers(timerFreq int64) {
	timerMultiplier = int64(refFreq / timerFreq)
	TimerFn = read_cntpct

	return
}

//go:linkname nanotime1 runtime.nanotime1
func nanotime1() int64 {
	return int64(TimerFn() * timerMultiplier)
}
