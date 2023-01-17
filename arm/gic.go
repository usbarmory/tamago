// ARM processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm

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
)

// InitGIC initializes the ARM Generic Interrupt Controller (GIC).
func (cpu *CPU) InitGIC(base uint32, secure bool) {
	cpu.gicd = base + GICD_OFF
	cpu.gicc = base + GICC_OFF

	// Get the maximum number of external interrupt lines
	itLinesNum := reg.Get(cpu.gicd+GICD_TYPER, GICD_TYPER_ITLINES, 0x1f)

	// Add a line for the 32 internal interrupts
	itLinesNum += 1

	for n := uint32(0); n < itLinesNum; n++ {
		// Disable interrupts
		addr := cpu.gicd + GICD_ICENABLER + 4*n
		reg.Write(addr, 0xffffffff)

		// Clear pending interrupts
		addr = cpu.gicd + GICD_ICPENDR + 4*n
		reg.Write(addr, 0xffffffff)

		if !secure {
			addr = cpu.gicd + GICD_IGROUPR + 4*n
			reg.Write(addr, 0xffffffff)
		}
	}

	// Set priority mask to allow Non-Secure world to use the lower half
	// of the priority range.
	reg.Write(cpu.gicc+GICC_PMR, 0x80)

	// Enable GIC
	reg.Write(cpu.gicc+GICC_CTLR, GICC_CTLR_ENABLEGRP1|GICC_CTLR_ENABLEGRP0|GICC_CTLR_FIQEN)
	reg.Set(cpu.gicd+GICD_CTLR, GICD_CTLR_ENABLEGRP1)
	reg.Set(cpu.gicd+GICD_CTLR, GICD_CTLR_ENABLEGRP0)
}

func irq(gicd uint32, m int, secure bool, enable bool) {
	if gicd == 0 {
		return
	}

	var addr uint32

	n := uint32(m / 32)
	i := m % 32

	if enable {
		addr = gicd + GICD_IGROUPR + 4*n

		if !secure {
			reg.Set(addr, i)
		} else {
			reg.Clear(addr, i)
		}

		addr = gicd + GICD_ISENABLER + 4*n
	} else {
		addr = gicd + GICD_ICENABLER + 4*n
	}

	reg.SetTo(addr, i, true)
}

// EnableInterrupt enables forwarding of the corresponding interrupt to the CPU
// and configures its group status (Secure: Group 0, Non-Secure: Group 1).
func (cpu *CPU) EnableInterrupt(id int, secure bool) {
	irq(cpu.gicd, id, secure, true)
}

// DisableInterrupt disables forwarding of the corresponding interrupt to the
// CPU.
func (cpu *CPU) DisableInterrupt(id int) {
	irq(cpu.gicd, id, false, false)
}

// GetInterrupt obtains and acknowledges a signaled interrupt.
func (cpu *CPU) GetInterrupt() (id int) {
	if cpu.gicc == 0 {
		return
	}

	return int(reg.Get(cpu.gicc + GICC_IAR, GICC_IAR_ID, 0x3ff))
}
