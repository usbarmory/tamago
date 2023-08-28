// ARM Generic Interrupt Controller (GIC) driver
// https://github.com/usbarmory/tamago
//
// IP: ARM Generic Interrupt Controller version 2.0
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package gic implements a driver for the ARM Generic Interrupt Controller.
//
// The driver is based on the following reference specifications:
//   - ARM IHI 0048B.b - ARM Generic Interrupt Controller - Architecture version 2.0
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/usbarmory/tamago.
package gic

import (
	"github.com/usbarmory/tamago/internal/reg"
)

const (
	// GIC offsets in Cortex-A7
	// (p178, Table 8-1, Cortex-A7 MPCore Technical Reference Manual).
	GICD_OFF = 0x1000
	GICC_OFF = 0x2000

	// Distributor register map
	// (p75, Table 4-1, ARM Generic Interrupt Controller Architecture Specification).
	GICD_CTLR            = 0x000
	GICD_CTLR_ENABLEGRP1 = 1
	GICD_CTLR_ENABLEGRP0 = 0

	GICD_TYPER         = 0x004
	GICD_TYPER_ITLINES = 0

	GICD_IGROUPR   = 0x080
	GICD_ISENABLER = 0x100
	GICD_ICENABLER = 0x180
	GICD_ICPENDR   = 0x280

	// CPU interface register map
	// (p76, Table 4-2, ARM Generic Interrupt Controller Architecture Specification).
	GICC_CTLR            = 0x0000
	GICC_CTLR_FIQEN      = 3
	GICC_CTLR_ENABLEGRP1 = 1
	GICC_CTLR_ENABLEGRP0 = 0

	GICC_PMR          = 0x0004
	GICC_PMR_PRIORITY = 0

	GICC_IAR    = 0x000c
	GICC_IAR_ID = 0

	GICC_EOIR    = 0x0010
	GICC_EOIR_ID = 0

	GICC_AIAR    = 0x0020
	GICC_AIAR_ID = 0

	GICC_AEOIR    = 0x0024
	GICC_AEOIR_ID = 0
)

// GIC represents the Generic Interrupt Controller instance.
type GIC struct {
	// Base register
	Base uint32

	// control registers
	gicd uint32
	gicc uint32
}

// InitGIC initializes the ARM Generic Interrupt Controller (GIC).
func (hw *GIC) Init(secure bool, fiqen bool) {
	if hw.Base == 0 {
		panic("invalid GIC instance")
	}

	hw.gicd = hw.Base + GICD_OFF
	hw.gicc = hw.Base + GICC_OFF

	// Get the maximum number of external interrupt lines
	itLinesNum := reg.Get(hw.gicd+GICD_TYPER, GICD_TYPER_ITLINES, 0x1f)

	// Add a line for the 32 internal interrupts
	itLinesNum += 1

	for n := uint32(0); n < itLinesNum; n++ {
		// Disable interrupts
		addr := hw.gicd + GICD_ICENABLER + 4*n
		reg.Write(addr, 0xffffffff)

		// Clear pending interrupts
		addr = hw.gicd + GICD_ICPENDR + 4*n
		reg.Write(addr, 0xffffffff)

		if !secure {
			addr = hw.gicd + GICD_IGROUPR + 4*n
			reg.Write(addr, 0xffffffff)
		}
	}

	// Set priority mask to allow Non-Secure world to use the lower half
	// of the priority range.
	reg.Write(hw.gicc+GICC_PMR, 0x80)

	reg.SetTo(hw.gicc+GICC_CTLR, GICC_CTLR_FIQEN, fiqen)

	reg.Set(hw.gicc+GICC_CTLR, GICC_CTLR_ENABLEGRP1)
	reg.Set(hw.gicc+GICC_CTLR, GICC_CTLR_ENABLEGRP0)

	reg.Set(hw.gicd+GICD_CTLR, GICD_CTLR_ENABLEGRP1)
	reg.Set(hw.gicd+GICD_CTLR, GICD_CTLR_ENABLEGRP0)
}

// FIQEn controls whether Group 0 (Secure) interrupts should be signalled as
// IRQ or FIQ requests.
func (hw *GIC) FIQEn(fiq bool) {
	if hw.gicc == 0 {
		return
	}

	reg.SetTo(hw.gicc+GICC_CTLR, GICC_CTLR_FIQEN, fiq)
}

func irq(gicd uint32, m int, secure bool, enable bool) {
	if gicd == 0 {
		return
	}

	var addr uint32

	n := uint32(m / 32)
	i := m % 32

	if enable {
		reg.SetTo(gicd + GICD_IGROUPR + 4*n, i, !secure)
		addr = gicd + GICD_ISENABLER + 4*n
	} else {
		addr = gicd + GICD_ICENABLER + 4*n
	}

	reg.SetTo(addr, i, true)
}

// EnableInterrupt enables forwarding of the corresponding interrupt to the CPU
// and configures its group status (Secure: Group 0, Non-Secure: Group 1).
func (hw *GIC) EnableInterrupt(id int, secure bool) {
	irq(hw.gicd, id, secure, true)
}

// DisableInterrupt disables forwarding of the corresponding interrupt to the
// CPU.
func (hw *GIC) DisableInterrupt(id int) {
	irq(hw.gicd, id, false, false)
}

// GetInterrupt obtains and acknowledges a signaled interrupt, the end of its
// handling must be signaled by closing the returned channel.
func (hw *GIC) GetInterrupt(secure bool) (id int, end chan struct{}) {
	if hw.gicc == 0 {
		return
	}

	var m uint32

	if secure {
		m = reg.Get(hw.gicc+GICC_IAR, GICC_IAR_ID, 0x3ff)
	} else {
		m = reg.Get(hw.gicc+GICC_AIAR, GICC_AIAR_ID, 0x3ff)
	}

	if m < 1020 {
		end = make(chan struct{})

		go func() {
			<-end

			if secure {
				reg.SetN(hw.gicc+GICC_EOIR, GICC_EOIR_ID, 0x3ff, m)
			} else {
				reg.SetN(hw.gicc+GICC_AEOIR, GICC_AEOIR_ID, 0x3ff, m)
			}
		}()
	}

	id = int(m)

	return
}
