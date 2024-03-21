// SiFive Core-Local Interruptor (CLINT) driver
// https://github.com/usbarmory/tamago
//
// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in Go LICENSE file.

package clint

func mulDiv(x, m, d uint64) uint64 {
	divx := x / d
	modx := x - divx*d
	divm := m / d
	modm := m - divm*d
	return divx*m + modx*divm + modx*modm/d
}

// Nanotime returns the number of nanoseconds counted from the RTCCLK input
// plus the timer offset.
func (hw *CLINT) Nanotime() int64 {
	return int64(mulDiv(hw.Mtime(), 1e9, hw.RTCCLK)) + hw.TimerOffset
}
