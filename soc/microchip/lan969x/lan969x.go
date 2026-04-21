// Microchip LAN969x configuration and support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package lan969x provides support to Go bare metal unikernels, written using
// the TamaGo framework, on the Microchip LAN969x family of System-on-Chip
// (SoC) application processors.
//
// The package implements initialization and drivers for Microchip LAN969x SoCs,
// adopting the following reference specifications:
//   - Microchip - LAN9694/LAN9696/LAN9698 Datasheet - DS00005048E (02-27-25)
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm64` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package lan969x

import (
	"github.com/usbarmory/tamago/arm/gic"
	"github.com/usbarmory/tamago/arm64"

	"github.com/usbarmory/tamago/soc/microchip/analyzer"
	"github.com/usbarmory/tamago/soc/microchip/flexcom"
	"github.com/usbarmory/tamago/soc/microchip/gpio"
	"github.com/usbarmory/tamago/soc/microchip/miim"
	"github.com/usbarmory/tamago/soc/microchip/otpc"
	"github.com/usbarmory/tamago/soc/microchip/trng"
)

// Ports defines the number of available Ethernet ports
const Ports = 30

// Interrupts
const (
	// CPU Device Queue System
	XTR_READY_IRQ = 42
	INJ_READY_IRQ = 43
)

// Peripheral registers
const (
	// Analyzer
	ANA_CL_BASE = 0xe2400000 // classifier
	ANA_L3_BASE = 0xe2480000 // layer 3 handling and routing
	ANA_L2_BASE = 0xe2800000 // layer 2 forwarding and learning
	ANA_AC_BASE = 0xe2900000 // access control

	// Assembler
	ASM_BASE = 0xe3200000

	// CPU system registers
	CPU_BASE = 0xe00c0000

	// DDR base address
	DDR_BASE = 0x60000000

	// CPU Device Queue System
	DEVCPU_QS = 0xe2030000

	// RGMII interfaces
	DEVRGMII0 = 0xe30e4000 // port 28
	DEVRGMII1 = 0xe30e8000 // port 29

	// Disassembler
	DSM_BASE = 0xe30ec000

	// Egress Access Control Lists
	EACL_BASE = 0xe22c0000

	// Serial ports
	FLEXCOM0_BASE = 0xe0040000
	FLEXCOM1_BASE = 0xe0044000
	FLEXCOM2_BASE = 0xe0060000
	FLEXCOM3_BASE = 0xe0064000

	// General Configuration Block
	GCB_BASE = 0xe2010000

	// General Interrupt Controller
	GIC_BASE = 0xe8c10000

	// General Purpose I/O Control
	GPIO_BASE = 0xe20100d4

	// Hierarchical Scheduler Configuration
	HSCH_BASE = 0xe2580000

	// High-speed I/O
	HSIO_BASE = 0xe3408000

	// Learn block
	LRN_BASE = 0xe2060000

	// PHY Management Controller
	MIIM0_BASE = 0xe20101a8
	MIIM1_BASE = 0xe20101cc

	// One Time Programmable Controller
	OTPC_BASE = 0xe0021000

	// DEVCPU Precision Timing Protocol Originator
	PTP_BASE = 0xe2040000

	// Queue Forwarding
	QFWD_BASE = 0xe20b0000

	// Queue System Configuration
	QSYS_BASE = 0xe20a0000

	// Rewriter
	REW_BASE = 0xe2600000

	// True Random Number Generator
	TRNG_BASE = 0xe0048000

	// Versatile Content Aware Processor
	VCAP_SUPER_BASE = 0xe2080000

	// Versatile OAM MEP Processor (VOP) block
	VOP_BASE = 0xe2a00000
)

// Peripheral instances
var (
	// Analyzer
	ANA = &analyzer.ANA{
		AccessControl: ANA_AC_BASE,
		Learn:         LRN_BASE,
	}

	// ARM64 core
	ARM64 = &arm64.CPU{
		// required before Init()
		TimerOffset: 1,
	}

	// Serial port 1
	FLEXCOM0 = &flexcom.FLEXCOM{
		Index: 1,
		Base:  FLEXCOM0_BASE,
	}

	// Generic Interrupt Controller
	GIC = &gic.GIC{
		Base: GIC_BASE,
	}

	// GPIO controller 1
	GPIO = &gpio.GPIO{
		Base: GPIO_BASE,
	}

	// PHY Management Controller 1
	MIIM0 = &miim.MIIM{
		Base: MIIM0_BASE,
	}

	// PHY Management Controller 2
	MIIM1 = &miim.MIIM{
		Base: MIIM1_BASE,
	}

	// One Time Programmable Controller
	OTPC = &otpc.OTPC{
		Base: OTPC_BASE,
		Size: 16*1024,
	}

	// True Random Number Generator
	TRNG = &trng.TRNG{
		Base: TRNG_BASE,
	}
)
