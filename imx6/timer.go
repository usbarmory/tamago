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
func read_cntpct() int64
func busyloop(int32)

// initGlobalTimers initializes ARM Cortex-A9 timers
func initGlobalTimers() {
	timerFn = read_gtc
	timerMultiplier = 10
}

// initGenericTimers initializes ARM Cortex-A7 timers
func initGenericTimers() {
	var timerFreq int64

	if !Native {
		// use QEMU fixed CNTFRQ value (62.5MHz)
		timerFreq = 62500000
	} else {
		// U-Boot value for i.MX6 family (8.0MHz)
		timerFreq = 8000000
	}

	timerMultiplier = int64(refFreq / timerFreq)
	timerFn = read_cntpct

	return
}

//go:linkname nanotime runtime.nanotime
func nanotime() int64 {
	return int64(timerFn() * timerMultiplier)
}
