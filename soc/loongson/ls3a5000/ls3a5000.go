// Loongson 3A5000/LS7A support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package ls3a5000 provides support for Loongson 3A5000/3A6000 processors
// paired with the LS7A bridge, as emulated by the QEMU LoongArch `virt`
// machine.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=loong64` as
// supported by the TamaGo framework for bare metal Go on LoongArch SoCs, see
// https://github.com/usbarmory/tamago.
package ls3a5000

import (
	"github.com/usbarmory/tamago/loong64"
	"github.com/usbarmory/tamago/soc/loongson/uart"
)

// Peripheral registers
const (
	// LS7A legacy NS16550 console, also exposed by the QEMU `virt` machine
	UART0_BASE = 0x1fe001e0
)

// Peripheral instances
var (
	// LA64 is the LoongArch 64-bit CPU instance
	LA64 = &loong64.CPU{
		// Counter must be set at package initialization as Nanotime
		// is invoked by the runtime before [Init].
		Counter:         loong64.Rdtime,
		TimerMultiplier: 1,
		TimerOffset:     1,
	}

	// UART0 is the LS7A legacy serial console
	UART0 = &uart.UART{
		Base: UART0_BASE,
	}
)

// Init takes care of the lower level SoC initialization triggered early in
// runtime setup.
func Init() {
	LA64.Init()
}
