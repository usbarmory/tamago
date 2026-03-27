// AI Foundry ET-SoC-1 Minion configuration and support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package minion provides support to Go bare metal unikernels, written using
// the TamaGo framework, on the AI Foundry ET-SoC-1 Minion Core.
//
// The package implements initialization and drivers for specific ET-SoC-1
// Minion Core peripherals, adopting, where indicated, the following
// reference specifications:
//   - ETSOC1RM - Esperanto SoC Programmer’s Reference Manual - Rev 1.0 2021/12
//
// This package is only meant to be used with `GOOS=tamago GOARCH=riscv64
// GOSOFT=1` as supported by the TamaGo framework for bare metal Go on RISC-V
// SoCs, see https://github.com/usbarmory/tamago.
package minion

import (
	"github.com/usbarmory/tamago/riscv64"
)

// Peripheral instances
var (
	// RISC-V core
	RV64 = &riscv64.CPU{
		Counter:         Counter,
		TimerMultiplier: TimerMultiplier,
		// required before Init()
		TimerOffset: 1,
	}
)

// Model returns the processor name.
func Model() (name string) {
	return "ET-SoC-1 Minion Core"
}
