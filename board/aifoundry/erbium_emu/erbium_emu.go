// AI Foundry Erbium emulator support for tamago/riscv64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package erbium_emu provides hardware initialization, automatically on
// import, for the AI Foundry Erbium emulator machine configured with a single
// core.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=riscv64` as
// supported by the TamaGo framework for bare metal Go on RISC-V SoCs, see
// https://github.com/usbarmory/tamago.
package erbium_emu

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/soc/aifoundry/erbium"
)

// Peripheral instances
var (
	UART0 = erbium.UART0
)

// Init takes care of the lower level initialization triggered early in runtime
// setup (post World start).
//
//go:linkname Init runtime/goos.Hwinit1
func Init() {
	// initialize SoC
	erbium.Init()

	// initialize serial console
	erbium.UART0.Init()
}
