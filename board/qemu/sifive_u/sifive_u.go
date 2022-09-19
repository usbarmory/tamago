// QEMU virt support for tamago/riscv64
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package sifive_u provides hardware initialization, automatically on import,
// for the QEMU sifive_u machine configured (see the `dts` file in this
// directory) with a single U54 core.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=riscv64` as
// supported by the TamaGo framework for bare metal Go on RISC-V SoCs, see
// https://github.com/usbarmory/tamago.
package sifive_u

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/soc/sifive/fu540"
)

// Peripheral instances
var (
	UART0 = fu540.UART0
)

// Init takes care of the lower level SoC initialization triggered early in
// runtime setup.
//
//go:linkname Init runtime.hwinit
func Init() {
	// initialize SoC
	fu540.Init()

	// initialize serial console
	fu540.UART0.Init()
}
