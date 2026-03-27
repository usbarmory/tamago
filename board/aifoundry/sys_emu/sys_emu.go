// AI Foundry ET-SoC-1 emulator support for tamago/riscv64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package sys_emu provides hardware initialization, automatically on import,
// for a single core running on the AI Foundry ET-SoC-1 emulator machine.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=riscv64` as
// supported by the TamaGo framework for bare metal Go on RISC-V SoCs, see
// https://github.com/usbarmory/tamago.
package sys_emu

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/soc/aifoundry/etsoc1"
	"github.com/usbarmory/tamago/soc/aifoundry/etsoc1/minion"
)

// Peripheral instances
var (
	UART0 = etsoc1.UART0
)

// Init takes care of the lower level initialization triggered early in runtime
// setup (post World start).
//
//go:linkname Init runtime/goos.Hwinit1
func Init() {
	// initialize ET-Minion core
	minion.Init()

	// initialize serial console
	etsoc1.UART0.Init()
}
