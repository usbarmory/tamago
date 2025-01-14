// microvm support for tamago/amd64
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
	"runtime"
	_ "unsafe"

	"github.com/usbarmory/tamago/amd64"
)

// Peripheral registers
const (
	COM1             = 0x3f8

	VIRTIO_MMIO_BASE = 0xfeb00000
	VIRTIO_NET_BASE  = VIRTIO_MMIO_BASE + 0x2e00
)

// Peripheral instances
var (
	// AMD64 core
	AMD64 = &amd64.CPU{}

	// Real-Time Clock
	RTC0 = &RTC{}

	// Serial port
	UART0 = &UART{
		Index: 1,
		Base:  COM1,
	}
)

//go:linkname nanotime1 runtime.nanotime1
func nanotime1() int64 {
	return int64(float64(AMD64.TimerFn())*AMD64.TimerMultiplier) + AMD64.TimerOffset
}

// Init takes care of the lower level initialization triggered early in runtime
// setup.
//
//go:linkname Init runtime.hwinit
func Init() {
	// initialize CPU
	AMD64.Init()

	// initialize serial console
	UART0.Init()

	runtime.Exit = func(_ int32) {
		// On microvm the recommended way to trigger a guest-initiated
		// shut down is by generating a triple-fault.
		amd64.Fault()
	}
}
