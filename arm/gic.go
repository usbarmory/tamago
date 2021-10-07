// ARM processor support
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm

import (
	"github.com/f-secure-foundry/tamago/internal/reg"
)

const (
	// GIC offsets in Cortex-A7
	// (p178, Table 8-1, Cortex-A7 MPCore Technical Reference Manual).
	GICD_OFF = 0x1000
	GICC_OFF = 0x2000

	// Distributor register map
	// (p75, Table 4-1, ARM Generic Interrupt Controller Architecture Specification).
	GICD_CTLR            = 0
	GICD_CTLR_ENABLEGRP1 = 1
	GICD_CTLR_ENABLEGRP0 = 0

	GICD_TYPER         = 0x4
	GICD_TYPER_ITLINES = 0

	GICD_IGROUPR   = 0x080
	GICD_ICENABLER = 0x180
	GICD_ICPENDR   = 0x280

	// CPU interface register map
	// (p76, Table 4-2, ARM Generic Interrupt Controller Architecture Specification).
	GICC_CTLR            = 0
	GICC_CTLR_FIQEN      = 3
	GICC_CTLR_ENABLEGRP1 = 1
	GICC_CTLR_ENABLEGRP0 = 0

	GICC_PMR          = 0x4
	GICC_PMR_PRIORITY = 0
)

// InitGIC initializes the ARM Generic Interrupt Controller (GIC).
func InitGIC(base uint32) {
	gicd := base + GICD_OFF
	gicc := base + GICC_OFF

	// Get the maximum number of external interrupt lines
	itLinesNum := reg.Get(gicd+GICD_TYPER, GICD_TYPER_ITLINES, 0x1f)

	// Add a line for the 32 internal interrupts
	itLinesNum += 1

	for i := uint32(0); i < itLinesNum; i++ {
		// Disable interrupts
		addr := gicd + GICD_ICENABLER + 4*i
		reg.Write(addr, 0xffffffff)

		// Clear pending interrupts
		addr = gicd + GICD_ICPENDR + 4*i
		reg.Write(addr, 0xffffffff)

		// Assign all interrupts to Non-Secure
		addr = GICD_IGROUPR + 4*i
		reg.Write(addr, 0xffffffff)
	}

	// Set priority mask to allow Non-Secure world to use the lower half
	// of the priority range.
	reg.Write(gicc+GICC_PMR, 0x80)

	// Enable GIC
	reg.Write(gicc+GICC_CTLR, GICC_CTLR_ENABLEGRP1|GICC_CTLR_ENABLEGRP0|GICC_CTLR_FIQEN)
	reg.Set(gicd+GICD_CTLR, GICD_CTLR_ENABLEGRP1)
	reg.Set(gicd+GICD_CTLR, GICD_CTLR_ENABLEGRP0)
}
