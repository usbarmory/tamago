// Nuvoton NUC980 SoC support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkramstart

package nuc980

import (
	_ "unsafe"
)

// ramStart is the base of DDR SDRAM on the NUC980.
// The ARM926EJ-S exception vectors are also placed here (fixed at 0x00000000).
//
//go:linkname ramStart runtime/goos.RamStart
var ramStart uint32 = 0x00000000
