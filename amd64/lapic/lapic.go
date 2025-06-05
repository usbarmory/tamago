// Intel Advanced Programmable Interrupt Controller (APIC) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package lapic implements a driver for the Intel Local (LAPIC) Advanced
// Programmable Interrupt Controllers adopting the following reference
// specifications:
//   - Intel® 64 and IA-32 Architectures Software Developer’s Manual - Volume 3A - Chapter 10
//
// This package is only meant to be used with `GOOS=tamago` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package lapic

import (
	"github.com/usbarmory/tamago/internal/reg"
)

// LAPIC registers
const (
	LAPICID = 0x20

	LAPICVER    = 0x30
	VER_ENTRIES = 16

	LAPICEOI = 0xb0

	LAPICSVR   = 0xf0
	SVR_ENABLE = 8
)

// LAPIC represents a Local APIC instance.
type LAPIC struct {
	// Base register
	Base uint32
}

// ID returns the LAPIC identification register.
func (io *LAPIC) ID() uint32 {
	return reg.Get(io.Base+LAPICID, 24, 0xf)
}

// Version returns the LAPIC version register.
func (io *LAPIC) Version() uint32 {
	return reg.Read(io.Base + LAPICVER)
}

// Entries returns the size of the LAPIC local vector table.
func (io *LAPIC) Entries() int {
	maxIndex := reg.Get(io.Base+LAPICVER, VER_ENTRIES, 0xff)
	return int(maxIndex) + 1
}

// Enable enables the Local APIC.
func (io *LAPIC) Enable() {
	reg.Set(io.Base+LAPICSVR, SVR_ENABLE)
}

// Disable disables the Local APIC.
func (io *LAPIC) Disable() {
	reg.Clear(io.Base+LAPICSVR, SVR_ENABLE)
}

// ClearInterrupt signals the end of an interrupt handling routine.
func (io *LAPIC) ClearInterrupt() {
	reg.Write(io.Base+LAPICEOI, 0)
}
