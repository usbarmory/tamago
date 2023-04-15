// SiFive Core-Local Interruptor (CLINT) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package clint implements a driver for SiFive Core-Local Interruptor (CLINT)
// block adopting the following reference specifications:
//   - FU540C00RM - SiFive FU540-C000 Manual - v1p4 2021/03/25
//
// This package is only meant to be used with `GOOS=tamago GOARCH=riscv64` as
// supported by the TamaGo framework for bare metal Go on RISC-V SoCs, see
// https://github.com/usbarmory/tamago.
package clint

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/internal/reg"
)

// CLINT registers
const (
	MTIME = 0xbff8
)

// CLINT represents a Core-Local Interruptor (CLINT) instance.
type CLINT struct {
	// Base register
	Base uint64
	// CPU real time clock
	RTCCLK uint64
	// Timer offset in nanoseconds
	TimerOffset int64
}

// Mtime returns the number of cycles counted from the RTCCLK input.
func (hw *CLINT) Mtime() uint64 {
	return reg.Read64(hw.Base + MTIME)
}

// SetTimer sets the timer to the argument nanoseconds value.
func (hw *CLINT) SetTimer(t int64) {
	hw.TimerOffset = t - hw.Nanotime()
}
