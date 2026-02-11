// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm64

import (
	_ "unsafe"
)

// Init takes care of the lower level initialization triggered before runtime
// setup (pre World start).
//
//go:linkname Init runtime/goos.Hwinit0
func Init() {
	fp_enable()

	// At start all memory is mapped as device memory, causing LDP/STP
	// instructions to require 8-byte alignment.
	//
	// To prevent faults, MMU initialization is done as soon as possible in
	// hwinit0, rather than in hwinit1.
	cpu := &CPU{}
	cpu.InitMMU()
}
