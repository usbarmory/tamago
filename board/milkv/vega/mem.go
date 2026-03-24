// MilkV Vega support for tamago/riscv64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkramsize && !qemu

package vega

import (
	_ "unsafe"
)

// Applications can override ramSize with the `linkramsize` build tag.

//go:linkname ramSize runtime/goos.RamSize
var ramSize uint64 = 0x0F000000 // 240 MB
