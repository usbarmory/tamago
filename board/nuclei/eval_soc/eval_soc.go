// Nuclei EvalSoC emulator support for tamago/riscv64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package eval_soc provides hardware initialization, automatically on import,
// for the Nuclei EvalSoC machine emulated by the Nuclei QEMU fork
// (qemu-system-riscv64 -M nuclei_evalsoc).
//
// The EvalSoC reuses the Fisilink FSL91030 SoC package (Nuclei UX600 core) and
// is closely related to the MilkV Vega board (board/milkv/vega): both run the
// same UX600 core and SiFive-compatible peripherals. This package adapts the
// FSL91030 SoC to the emulator, which does not model the GPIO block or the
// hardware CLINT timer, supplying a QEMU-compatible memory map and time source.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=riscv64` as
// supported by the TamaGo framework for bare metal Go on RISC-V SoCs, see
// https://github.com/usbarmory/tamago.
package eval_soc

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
	// initialize the RISC-V core; the emulator does not model the GPIO
	// block, so UART pinmux (required only on real hardware) is skipped.
	fsl91030.RV64.Init()

	// initialize serial console
	fsl91030.UART0.Init()
}
