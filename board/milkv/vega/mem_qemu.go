// MilkV Vega support for tamago/riscv64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkramsize && qemu

package vega

import (
	_ "unsafe"
)

// QEMU nuclei_evalsoc limits RAM to 200 MB to avoid an internal memory
// region conflict in the emulator. The full 240 MB (DRAM end 0x50000000)
// causes overlap with QEMU's nuclei_evalsoc internal address decoding.

//go:linkname ramSize runtime/goos.RamSize
var ramSize uint64 = 0x0C800000 // 200 MB
