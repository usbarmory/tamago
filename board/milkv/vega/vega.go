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

	"github.com/usbarmory/tamago/soc/fisilink/fsl91030"
)

// Peripheral instances
var (
	UART0 = fsl91030.UART0
	UART1 = fsl91030.UART1
)

// Init takes care of the lower level initialization triggered early in runtime
// setup (post World start).
//
//go:linkname Init runtime/goos.Hwinit1
func Init() {
	// initialize SoC (CPU, GPIO pinmux)
	fsl91030.Init()

	// initialize serial console
	fsl91030.UART0.Init()
}
