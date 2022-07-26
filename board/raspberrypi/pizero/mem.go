// Raspberry Pi Zero support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) the pizero package authors
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkramsize
// +build !linkramsize

package pizero

import (
	_ "unsafe"
)

//go:linkname ramSize runtime.ramSize
var ramSize uint32 = 0x20000000 - 0x04000000 // 512 MB - 64MB GPU (VideoCore)
