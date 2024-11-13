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

	_ "github.com/usbarmory/tamago/amd64"
)

// Init takes care of the lower level initialization triggered early in runtime
// setup.
//
//go:linkname Init runtime.hwinit
func Init() {
	// FIXME: TODO
}
