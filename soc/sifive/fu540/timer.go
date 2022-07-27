// SiFive FU540 timer support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package fu540

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/internal/reg"
)

var timerOffset int64

func mulDiv(x, m, d uint64) uint64 {
	divx := x / d
	modx := x - divx*d
	divm := m / d
	modm := m - divm*d
	return divx*m + modx*divm + modx*modm/d
}

//go:linkname nanotime1 runtime.nanotime1
func nanotime1() int64 {
	v := reg.Read64(CLINT_BASE + MTIME)
	return int64(mulDiv(v, 1e9, RTCCLK)) + timerOffset
}

// SetTimer sets the timer to the argument nanoseconds value.
func SetTimer(t int64) {
	timerOffset = t - nanotime1()
}
