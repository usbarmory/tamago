// Raspberry Pi 2 support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) the pi2 package authors
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkramsize
// +build !linkramsize

package pi2

import (
	_ "unsafe"
)

//go:linkname ramSize runtime.ramSize
var ramSize uint32 = 0x40000000 - 0x4C00000 // 1GB - 76MB (VideoCore)
