// QEMU LoongArch virt support for tamago/loong64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package virt provides hardware initialization, automatically on import, for
// the QEMU LoongArch `virt` machine.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=loong64` as
// supported by the TamaGo framework for bare metal Go on LoongArch SoCs, see
// https://github.com/usbarmory/tamago.
package virt

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/soc/loongson/ls3a5000"
)

// Peripheral instances
var (
	UART0 = ls3a5000.UART0
)

// Init takes care of the lower level initialization triggered early in runtime
// setup (post World start).
//
//go:linkname Init runtime/goos.Hwinit1
func Init() {
	// initialize SoC
	ls3a5000.Init()

	// initialize serial console
	ls3a5000.UART0.Init()
}
