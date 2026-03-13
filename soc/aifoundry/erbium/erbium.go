// AI Foundry Erbium configuration and support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package erbium provides support to Go bare metal unikernels, written using
// the TamaGo framework, on the AI Foundry Erbium processor.
//
// The package implements initialization and drivers for specific Erbium
// processor peripherals, adopting, where indicated, the following reference
// specifications:
//   - Erbium Minion Isotope TRM - 2025/12/03
//
// This package is only meant to be used with `GOOS=tamago GOARCH=riscv64` as
// supported by the TamaGo framework for bare metal Go on RISC-V SoCs, see
// https://github.com/usbarmory/tamago.
package erbium

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/riscv64"
	"github.com/usbarmory/tamago/soc/aifoundry/uart"
)

// Peripheral registers
const (
	// System configuration
	SYSTEM_CONFIG = 0x02000008

	// Machine Timer
	ESR_MTIME = 0x80f40200

	// Serial ports
	UART0_BASE = 0x02004000
)

const TimerMultiplier = 50000

// Peripheral instances
var (
	// RISC-V core
	RV64 = &riscv64.CPU{
		Counter:         Counter,
		TimerMultiplier: TimerMultiplier,
		// required before Init()
		TimerOffset: 1,
	}

	// Serial port 1
	UART0 = &uart.UART{
		Index:  1,
		Base:   UART0_BASE,
		System: SYSTEM_CONFIG,
	}
)

// Model returns the processor name.
func Model() string {
	return "Erbium"
}
