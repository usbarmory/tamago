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

package imx6

import (
	_ "unsafe"
)

// nanoseconds
const refFreq int64 = 1000000000

var timerFn func() int64
var timerMultiplier int64

// defined in timer_arm.s
func read_gtc() int64
func read_cntfrq() int32
func write_cntfrq(int32)
func read_cntpct() int64
func busyloop(int32)

// initGlobalTimers initializes ARM Cortex-A9 timers
func initGlobalTimers() {
	timerFn = read_gtc
	timerMultiplier = 10
}

// initGenericTimers initializes ARM Cortex-A7 timers
func initGenericTimers(freq int32) {
	var timerFreq int64

	if features.genericTimer && freq != 0 {
		write_cntfrq(freq)
	}

	timerFreq = int64(read_cntfrq())
	print("System counter frequency (CNTFRQ) = ", timerFreq, "\n")
	timerMultiplier = int64(refFreq / timerFreq)
	timerFn = read_cntpct

	return
}

//go:linkname nanotime1 runtime.nanotime1
func nanotime1() int64 {
	return int64(timerFn() * timerMultiplier)
}
