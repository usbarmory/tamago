// QEMU LoongArch virt support for tamago/loong64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkramsize

package virt

import (
	_ "unsafe"
)

// Applications can override ramSize with the `linkramsize` build tag.
//
//go:linkname ramSize runtime/goos.RamSize
var ramSize uint64 = 0x10000000 // 256MB
