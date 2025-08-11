// Intel Advanced Programmable Interrupt Controller (APIC) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package ioapic implements a driver for the Intel I/O (IOAPIC) Advanced
// Programmable Interrupt Controllers adopting the following reference
// specifications:
//   - 82093AA I/O Advanced Programmable Interrupt Controller (IOAPIC)
//
// This package is only meant to be used with `GOOS=tamago` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package ioapic

import (
	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
)

// I/O APIC supported vectors
const (
	MinVector = 16
	MaxVector = 255
)

// I/O APIC registers
const (
	IOREGSEL = 0x00
	IOWIN    = 0x10

	IOAPICID = 0x00

	IOAPICVER   = 0x01
	VER_ENTRIES = 16

	IOAPICREDTBLn  = 0x10
	REDTBL_DEST    = 56
	REDTBL_MASK    = 16
	REDTBL_DESTMOD = 11
	REDTBL_INTVEC  = 0
)

// IOAPIC represents an I/O APIC instance.
type IOAPIC struct {
	// Controller index
	Index int
	// Base register
	Base uint32
	// Global System Interrupt Base
	GSIBase int
}

// Init initializes the I/O APIC.
func (io *IOAPIC) Init() {
	reg.Write(io.Base+IOREGSEL, IOAPICID)
	reg.SetN(io.Base+IOWIN, 24, 0xf, uint32(io.Index))
}

// ID returns the IOAPIC identification.
func (io *IOAPIC) ID() uint32 {
	reg.Write(io.Base+IOREGSEL, IOAPICID)
	return reg.Get(io.Base+IOWIN, 24, 0xf)
}

// Version returns the IOAPIC version register.
func (io *IOAPIC) Version() uint32 {
	reg.Write(io.Base+IOREGSEL, IOAPICVER)
	return reg.Read(io.Base + IOWIN)
}

// Entries returns the size of the IOAPIC redirection table.
func (io *IOAPIC) Entries() int {
	reg.Write(io.Base+IOREGSEL, IOAPICVER)
	maxIndex := reg.Get(io.Base+IOWIN, VER_ENTRIES, 0xff)
	return int(maxIndex) + 1
}

// EnableInterrupt activates an IOAPIC redirection table entry at the
// corresponding index for the desired interrupt vector.
func (io *IOAPIC) EnableInterrupt(index int, id int) {
	var val uint32

	if id < MinVector || id > MaxVector {
		return
	}

	index -= io.GSIBase

	if index > io.Entries()-1 {
		return
	}

	// set destination field for physical mode
	bits.Clear(&val, REDTBL_DESTMOD)
	// set destination to BSP
	bits.SetN(&val, REDTBL_DEST, 0xf, 0)

	// set interrupt vector
	bits.Clear(&val, REDTBL_MASK)
	bits.SetN(&val, REDTBL_INTVEC, 0xff, uint32(id))

	// set redirection table entry
	reg.Write(io.Base+IOREGSEL, IOAPICREDTBLn+uint32(index*2))
	reg.Write(io.Base+IOWIN, val)
}
