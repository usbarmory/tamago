// Nuvoton Enhanced Timer (ETimer) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package etimer implements a driver for the Enhanced Timer (ETimer) blocks
// found on Nuvoton SoCs adopting the following reference specifications:
//   - NUC980 Series Datasheet - Rev 1.24
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package etimer

import (
	"github.com/usbarmory/tamago/internal/reg"
)

// ETimer register offsets (from ETimer.Base).
const (
	CTL    = 0x00 // Timer Control Register
	PRECNT = 0x04 // Pre-scale Counter Register
	CMPR   = 0x08 // Compare Register
	IER    = 0x0c // Interrupt Enable Register
	ISR    = 0x10 // Interrupt Status Register
	DR     = 0x14 // Data Register (current count)
)

// CTL register bits.
const (
	CTL_EN         = 0x01 // bit 0: timer enable (CEN)
	CTL_DBGACK     = 0x08 // bit 3: continue counting during ICE debug halt
	CTL_PERIODIC   = 0x10 // bits [5:4] = 01: periodic mode
	CTL_TOGGLE     = 0x20 // bits [5:4] = 10: toggle mode
	CTL_CONTINUOUS = 0x30 // bits [5:4] = 11: continuous (free-running)
)

// IER/ISR register bits.
const IER_CMP_IEN = 0 // bit 0: compare-match interrupt enable

// CMPR_MAX is the maximum value of the 24-bit compare/data register.
const CMPR_MAX = 0x00ffffff

// ETimer represents an Enhanced Timer instance.
type ETimer struct {
	// Base register
	Base uint32
}

// Stop disables the timer.
func (hw *ETimer) Stop() {
	reg.Write(hw.Base+CTL, 0)
}

// SetPrescale sets the pre-scale divider; the tick rate is the eclk source
// divided by (prescale + 1).
func (hw *ETimer) SetPrescale(prescale uint32) {
	reg.Write(hw.Base+PRECNT, prescale)
}

// SetCompare sets the 24-bit compare value at which the timer wraps or
// triggers its interrupt.
func (hw *ETimer) SetCompare(compare uint32) {
	reg.Write(hw.Base+CMPR, compare&CMPR_MAX)
}

// EnableInterrupt enables or disables the compare-match interrupt.
func (hw *ETimer) EnableInterrupt(enable bool) {
	reg.SetTo(hw.Base+IER, IER_CMP_IEN, enable)
}

// ClearInterrupt clears a pending compare-match interrupt.
func (hw *ETimer) ClearInterrupt() {
	reg.Write(hw.Base+ISR, 0x1)
}

// Start enables the timer with CTL_EN set together with the given mode bits
// (e.g. CTL_PERIODIC, CTL_DBGACK).
func (hw *ETimer) Start(mode uint32) {
	reg.Write(hw.Base+CTL, CTL_EN|mode)
}

// Count returns the current 24-bit counter value.
func (hw *ETimer) Count() uint32 {
	return reg.Read(hw.Base+DR) & CMPR_MAX
}

// Control returns the CTL register value.
func (hw *ETimer) Control() uint32 {
	return reg.Read(hw.Base + CTL)
}

// Compare returns the CMPR register value.
func (hw *ETimer) Compare() uint32 {
	return reg.Read(hw.Base + CMPR)
}

// InterruptEnabled returns the IER register value.
func (hw *ETimer) InterruptEnabled() uint32 {
	return reg.Read(hw.Base + IER)
}

// InterruptStatus returns the ISR register value.
func (hw *ETimer) InterruptStatus() uint32 {
	return reg.Read(hw.Base + ISR)
}
