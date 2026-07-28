// MilkV Vega support for tamago/riscv64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package vega provides hardware initialization, automatically on import,
// for the MilkV Vega board equipped with the Fisilink FSL91030 SoC
// (Nuclei UX600 RV64IMAFDC core, 240 MB DRAM at 0x41000000).
//
// This package is only meant to be used with `GOOS=tamago GOARCH=riscv64` as
// supported by the TamaGo framework for bare metal Go on RISC-V SoCs, see
// https://github.com/usbarmory/tamago.
package vega

import (
	_ "unsafe"

	fsl "github.com/usbarmory/tamago/soc/fisilink/fsl91030"
)

// UART0 signals are routed to GPIO pin 16 (TX) and pin 17 (RX) via IOF0
// connected to the onboard FTDI channel 0.
const (
	GPIO_UART0_TX = 16
	GPIO_UART0_RX = 17
)

// Peripheral instances
var (
	// UART0 is the USB-C console
	UART0 = fsl.UART0
	// UART1 is the rear connector
	UART1 = fsl.UART1
)

// Init takes care of the lower level initialization triggered early in runtime
// setup (post World start).
//
//go:linkname Init runtime/goos.Hwinit1
func Init() {
	// initialize SoC (CPU, GPIO pinmux)
	fsl.Init()

	// route UART0 signals to physical pins
	fsl.ConfigureGPIO(GPIO_UART0_TX, true)
	fsl.ConfigureGPIO(GPIO_UART0_RX, true)

	// initialize serial console
	fsl.UART0.Init()
}
