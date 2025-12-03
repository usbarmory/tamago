// ARM64 Generic Interrupt Controller (GICv3) driver
// https://github.com/usbarmory/tamago
//
// IP: ARM Generic Interrupt Controller version 3.0
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package gic implements a driver for the ARM Generic Interrupt Controller
// (GICv3).
//
// The driver is based on the following reference specifications:
//   - ARM IHI 0069G - ARM GIC Architecture Specification (v3 and v4)
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package gic

import (
	"time"

	"github.com/usbarmory/tamago/internal/reg"
)

// GIC Distributor register map
// (p519, Table 12-25 Distributor register map, ARM IHI 0069G).
const (
	GICD_CTLR       = 0x000
	CTLR_ARE_NS     = 5
	CTLR_ARE_S      = 4
	CTLR_ENABLEGRP0 = 0

	GICD_TYPER    = 0x004
	TYPER_ITLINES = 0

	GICD_IGROUPR   = 0x0080
	GICD_ISENABLER = 0x0100
	GICD_ICENABLER = 0x0180
	GICD_ICPENDR   = 0x0280
	GICD_IROUTER   = 0x6100
)

// GIC Redistributor register map
// (p615, Table 12-27 Redistributor register map, ARM IHI 0069G).
const (
	// Redistributor frames
	RD_BASE  = 0x00000
	SGI_BASE = 0x10000

	GICR_WAKER            = RD_BASE + 0x0014
	WAKER_CHILDREN_ASLEEP = 2
	WAKER_PROCESSOR_SLEEP = 1

	GICR_IGROUPR = SGI_BASE + 0x0080
)

const (
	firstSGI = 0    // Software Generated Interrupts (SGI)
	firstPPI = 16   // Private Peripheral Interrupts (PPI)
	firstSPI = 32   //  Shared Peripheral Interrupts (SPI)
	firstSIN = 1020 //     Special Interrupt Numbers
)

// GIC represents a Generic Interrupt Controller (GICv3) instance.
type GIC struct {
	// GIC Distributor base address
	GICD uint32
	// GIC Redistributor base address
	GICR uint32

	// core identifier
	mpidr uint64
}

// defined in gic.s
func write_icc_sre_el3(val uint64)
func write_icc_igrpen0_el1(val uint64)
func write_icc_pmr_el1(val uint64)
func read_icc_iar0() uint64
func read_mpidr_el1() uint64
func write_icc_eoir0(val uint64)

// InitGIC initializes an ARM Generic Interrupt Controller (GICv3) instance.
func (hw *GIC) Init() {
	if hw.GICD == 0 || hw.GICR == 0 {
		panic("invalid GIC instance")
	}

	// Mark CPU as being online
	reg.Clear(hw.GICR+GICR_WAKER, WAKER_PROCESSOR_SLEEP)

	if !reg.WaitFor(1*time.Second, hw.GICR+GICR_WAKER, WAKER_CHILDREN_ASLEEP, 1, 0) {
		panic("could not wake GICR")
	}

	// Get the maximum number of external interrupt lines
	itLinesNum := reg.Get(hw.GICD+GICD_TYPER, TYPER_ITLINES, 0x1f)

	// Add a line for the 32 internal interrupts
	itLinesNum += 1

	for n := uint32(0); n < itLinesNum; n++ {
		// Disable interrupts
		addr := hw.GICD + GICD_ICENABLER + 4*n
		reg.Write(addr, 0xffffffff)

		// Clear pending interrupts
		addr = hw.GICD + GICD_ICPENDR + 4*n
		reg.Write(addr, 0xffffffff)
	}

	// Enable system register interface
	write_icc_sre_el3(1)

	// Unmask all interrupt priorities
	write_icc_pmr_el1(0xff)

	// Enable Group0 interrupts (CPU interface)
	write_icc_igrpen0_el1(1)

	// Enable Group0 interrupts (Distributor)
	reg.Set(hw.GICD+GICD_CTLR, CTLR_ENABLEGRP0)

	// Enable affinity routing
	reg.Set(hw.GICD+GICD_CTLR, CTLR_ARE_NS)
	reg.Set(hw.GICD+GICD_CTLR, CTLR_ARE_S)

	// Cache core identifier
	hw.mpidr = read_mpidr_el1()
}

func (hw *GIC) irq(m int, enable bool) {
	if hw.GICD == 0 {
		return
	}

	var off uint32
	n := uint32(m / 32)
	i := m % 32

	if enable {
		if m < firstSPI {
			reg.Clear(hw.GICR+GICR_IGROUPR+4*n, i)
		} else {
			// route to core identified at initialization
			reg.Write64(uint64(hw.GICD+GICD_IROUTER)+uint64(8*m), hw.mpidr)
			// assign to Group0
			reg.Clear(hw.GICD+GICD_IGROUPR+4*n, i)
		}

		off += GICD_ISENABLER
	} else {
		off += GICD_ICENABLER
	}

	if m < firstSPI {
		reg.SetTo(hw.GICR+SGI_BASE+off+4*n, i, true)
	} else {
		reg.SetTo(hw.GICD+off+4*n, i, true)
	}
}

// EnableInterrupt enables forwarding of the corresponding interrupt to the CPU
// and configures its group status (Secure: Group 0, Non-Secure: Group 1).
func (hw *GIC) EnableInterrupt(id int) {
	hw.irq(id, true)
}

// DisableInterrupt disables forwarding of the corresponding interrupt to the
// CPU.
func (hw *GIC) DisableInterrupt(id int) {
	hw.irq(id, false)
}

// GetInterrupt obtains and acknowledges a signaled interrupt.
func (hw *GIC) GetInterrupt() (id int) {
	if hw.GICD == 0 {
		return
	}

	m := read_icc_iar0() & 0xffffff

	if m < firstSIN {
		write_icc_eoir0(m)
	}

	return int(m)
}
