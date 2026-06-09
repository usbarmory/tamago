// Nuvoton Advanced Interrupt Controller (AIC) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package aic implements a driver for the Advanced Interrupt Controller (AIC)
// found on Nuvoton SoCs adopting the following reference specifications:
//   - NUC980 Series Datasheet - Rev 1.24
//
// The AIC supports 64 interrupt sources (IRQ0..IRQ63). Sources 0..31 are
// controlled by the low-word registers, sources 32..63 by the high-word
// (H) registers.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package aic

import (
	"github.com/usbarmory/tamago/internal/reg"
)

// AIC register offsets (from AIC.Base).
const (
	SISCR = 0x100 // Software Interrupt Set Command Register
	SICCR = 0x104 // Software Interrupt Clear Command Register
	IMR   = 0x128 // Interrupt Mask Register
	ISNR  = 0x120 // Interrupt Source Number Register (current IRQ)
	MECR  = 0x130 // Mask Enable  Command Register, IRQ0..31
	MECRH = 0x134 // Mask Enable  Command Register, IRQ32..63
	MDCR  = 0x138 // Mask Disable Command Register, IRQ0..31
	MDCRH = 0x13c // Mask Disable Command Register, IRQ32..63
	EOSCR = 0x150 // End-of-Service Command Register
)

// maxIRQ is the highest valid interrupt source number.
const maxIRQ = 63

// AIC represents an Advanced Interrupt Controller instance.
type AIC struct {
	// Base register
	Base uint32
}

// DisableAll masks all 64 interrupt sources.
func (hw *AIC) DisableAll() {
	reg.Write(hw.Base+MDCR, 0xffffffff)
	reg.Write(hw.Base+MDCRH, 0xffffffff)
}

// EnableIRQ unmasks interrupt source irq (0..63).
func (hw *AIC) EnableIRQ(irq int) {
	if irq < 0 || irq > maxIRQ {
		return
	}
	if irq < 32 {
		reg.Write(hw.Base+MECR, 1<<uint(irq))
	} else {
		reg.Write(hw.Base+MECRH, 1<<uint(irq-32))
	}
}

// DisableIRQ masks interrupt source irq (0..63).
func (hw *AIC) DisableIRQ(irq int) {
	if irq < 0 || irq > maxIRQ {
		return
	}
	if irq < 32 {
		reg.Write(hw.Base+MDCR, 1<<uint(irq))
	} else {
		reg.Write(hw.Base+MDCRH, 1<<uint(irq-32))
	}
}

// CurrentIRQ returns the interrupt source number currently being serviced.
func (hw *AIC) CurrentIRQ() int {
	return int(reg.Read(hw.Base+ISNR) & 0x7f)
}

// EOI signals end-of-interrupt to the AIC.
func (hw *AIC) EOI() {
	reg.Write(hw.Base+EOSCR, 0x1)
}

// SoftwareInterrupt generates a software interrupt for source irq. The AIC
// SISCR register only covers IRQ0..31; higher sources are not supported.
func (hw *AIC) SoftwareInterrupt(irq int) {
	if irq < 0 || irq > 31 {
		return
	}
	reg.Write(hw.Base+SISCR, 1<<uint(irq))
}
