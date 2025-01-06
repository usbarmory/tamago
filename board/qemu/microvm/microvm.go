// QEMU microvm support for tamago/amd64
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package microvm provides hardware initialization, automatically on import,
// for the QEMU microvm machine configured with a single x86_64 core.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=amd64` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package microvm

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/amd64"
)

const COM1 = 0x3f8

// Peripheral instances
var (
	// AMD64 core
	AMD64 = &amd64.CPU{}

	// legacy serial console
	UART0 = &UART{
		Index: 1,
		Base:  COM1,
	}
)

// Init takes care of the lower level initialization triggered early in runtime
// setup.
//
//go:linkname Init runtime.hwinit
func Init() {
	// initialize CPU
	AMD64.Init()

	// initialize serial console
	UART0.Init()
}
