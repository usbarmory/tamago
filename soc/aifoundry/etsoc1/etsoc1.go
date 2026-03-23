// AI Foundry ET-SoC-1 configuration and support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package etsoc1 provides support to Go bare metal unikernels, written using
// the TamaGo framework, on the AI Foundry ET-SoC-1 processor.
//
// The package implements initialization and drivers for specific ET-SoC-1
// processor peripherals, adopting, where indicated, the following reference
// specifications:
//   - ETSOC1RM - Esperanto SoC Programmer’s Reference Manual - Rev 1.0 2021/12
//
// This package is only meant to be used with `GOOS=tamago GOARCH=riscv64` as
// supported by the TamaGo framework for bare metal Go on RISC-V SoCs, see
// https://github.com/usbarmory/tamago.
package etsoc1

import (
	"github.com/usbarmory/tamago/soc/aifoundry/uart"
)

// Peripheral registers
const (
	// External memory
	DRAM_BASE = 0x80_0000_0000

	// Serial ports
	UART0_BASE = 0x00_1200_2000
	UART1_BASE = 0x00_1200_7000
)

// Peripheral instances
var (
	// Serial port 0
	UART0 = &uart.Synopsys{
		Index: 0,
		Base:  UART0_BASE,
	}

	// Serial port 1
	UART1 = &uart.Synopsys{
		Index: 1,
		Base:  UART1_BASE,
	}
)
