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

	"github.com/usbarmory/tamago/soc/microchip/flexcom"
	"github.com/usbarmory/tamago/soc/microchip/trng"
)

// Peripheral registers
const (
	// DDR base address
	DDR_BASE = 0x60000000

	// Serial ports
	FLEXCOM0_BASE = 0xe0040200

	// General Interrupt Controller
	GIC_BASE = 0xe8c11000

	// True Random Number Generator
	TRNG_BASE = 0xe0048000
)

// Peripheral instances
var (
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

	// True Random Number Generator
	TRNG = &trng.TRNG{
		Base: TRNG_BASE,
	}
)
