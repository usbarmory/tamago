// Intel Advanced Programmable Interrupt Controller (APIC) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
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
	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
)

// LAPIC registers
const (
	LAPIC_ID = 0x20
	ID       = 24

	LAPIC_VER   = 0x30
	VER_ENTRIES = 16

	LAPIC_EOI = 0xb0

	LAPIC_SVR  = 0xf0
	SVR_ENABLE = 8

	LAPIC_ICRL = 0x300
	LAPIC_ICRH = 0x310

	ICR_DST      = 18
	ICR_DST_SELF = 0b01 << ICR_DST
	ICR_DST_ALL  = 0b10 << ICR_DST
	ICR_DST_REST = 0b11 << ICR_DST

	ICR_INIT       = 14
	ICR_DLV_STATUS = 12
	ICR_DLV        = 8

	ICR_DLV_SIPI = 0b110 << ICR_DLV
	ICR_DLV_INIT = 0b101 << ICR_DLV
	ICR_DLV_NMI  = 0b100 << ICR_DLV
	ICR_DLV_SMI  = 0b010 << ICR_DLV
	ICR_DLV_LOW  = 0b001 << ICR_DLV
	ICR_DLV_IRQ  = 0b000 << ICR_DLV

	LAPIC_LVT_TIMER = 0x320
	TIMER_MODE      = 17
	TIMER_IRQ       = 0

	TIMER_MODE_ONE_SHOT     = 0b00
	TIMER_MODE_PERIODIC     = 0b01
	TIMER_MODE_TSC_DEADLINE = 0b10
)

// LAPIC represents a Local APIC instance.
type LAPIC struct {
	// Base register
	Base uint32
}

// ID returns the LAPIC identification register.
func (io *LAPIC) ID() uint32 {
	return reg.GetN(io.Base+LAPIC_ID, ID, 0xf)
}

// Version returns the LAPIC version register.
func (io *LAPIC) Version() uint32 {
	return reg.Read(io.Base + LAPIC_VER)
}

// Entries returns the size of the LAPIC local vector table.
func (io *LAPIC) Entries() int {
	maxIndex := reg.GetN(io.Base+LAPIC_VER, VER_ENTRIES, 0xff)
	return int(maxIndex) + 1
}

// Enable enables the Local APIC.
func (io *LAPIC) Enable() {
	reg.Set(io.Base+LAPIC_SVR, SVR_ENABLE)
}

// Disable disables the Local APIC.
func (io *LAPIC) Disable() {
	reg.Clear(io.Base+LAPIC_SVR, SVR_ENABLE)
}

// ClearInterrupt signals the end of an interrupt handling routine.
func (io *LAPIC) ClearInterrupt() {
	reg.Write(io.Base+LAPIC_EOI, 0)
}

// IPI sends an Inter-Processor Interrupt (IPI).
func (io *LAPIC) IPI(apicid int, id int, flags int) {
	reg.SetN(io.Base+LAPIC_ICRH, ID, 0xff, uint32(apicid))
	reg.Write(io.Base+LAPIC_ICRL, uint32(flags&0xffffff00)|uint32(id&0xff))
	reg.Wait(io.Base+LAPIC_ICRL, ICR_DLV_STATUS, 1, 0)
}

// SetTimer configures the LAPIC LVT Timer with the argument vector and mode.
func (io *LAPIC) SetTimer(id int, mode int) {
	var val uint32

	bits.SetN(&val, TIMER_IRQ, 0xff, uint32(id))
	bits.SetN(&val, TIMER_MODE, 0b11, uint32(mode))

	reg.Write(io.Base+LAPIC_LVT_TIMER, val)
}
