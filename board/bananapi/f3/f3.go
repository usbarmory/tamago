// Banana Pi BPI-F3 support for tamago/riscv64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package f3 provides hardware initialization, automatically on import, for
// the Banana Pi BPI-F3 board, an single board computer
// equipped with the SpacemiT K1 SoC (eight RV64GCVB SpacemiT X60 cores) and
// LPDDR4x DRAM.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=riscv64` as
// supported by the TamaGo framework for bare metal Go on RISC-V SoCs, see
// https://github.com/usbarmory/tamago.
package f3

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/soc/spacemit/k1"
)

// UART0 signals are routed to the pads controlled by GPIO_68 (UART0_TXD) and
// GPIO_69 (UART0_RXD) through their alternate function 2 (p125, Section 4.5
// Multi-Function I/O Pin Assignments, K1 Datasheet), which are wired to the
// board 40-pin GPIO header debug console.
const (
	PAD_UART0_TX = 68
	PAD_UART0_RX = 69
)

// Peripheral instances
var (
	// UART0 is the debug console
	UART0 = k1.UART0
)

// Init takes care of the lower level initialization triggered early in runtime
// setup (post World start).
//
//go:linkname Init runtime/goos.Hwinit1
func Init() {
	// initialize SoC
	k1.Init()

	// route UART0 signals to physical pins
	k1.ConfigureGPIO(PAD_UART0_TX, k1.AF2)
	k1.ConfigureGPIO(PAD_UART0_RX, k1.AF2)

	// initialize serial console
	k1.UART0.Init()
}
