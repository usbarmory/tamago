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
// This package is only meant to be used with `GOOS=tamago GOARCH=riscv64
// GOSOFT=1` as supported by the TamaGo framework for bare metal Go on RISC-V
// SoCs, see https://github.com/usbarmory/tamago.
package erbium

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/riscv64"
	"github.com/usbarmory/tamago/soc/aifoundry/uart"
)

// Peripheral registers
const (
	// System register
	SYSREG_BASE = 0x02000000

	// System configuration
	SYSTEM_CONFIG = 0x02000008

	// Serial ports
	UART0_BASE = 0x02004000
)

// Peripheral instances
var (
	// ET-Minion RISC-V core
	RV64 = &riscv64.CPU{
		Counter:         Counter,
		TimerMultiplier: TimerMultiplier,
		// required before Init()
		TimerOffset: 1,
	}

	// Serial port
	UART0 = &uart.Shakti{
		Base:   UART0_BASE,
		System: SYSTEM_CONFIG,
	}
)

// Model returns the processor name and version.
func Model() (name string, version uint32) {
	return "Erbium", reg.Read(SYSREG_BASE)
}
