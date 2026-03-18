// NuMaker-IIoT-NUC980G2 board support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkramsize

package nuc980iiot

import (
	_ "unsafe"
)

// ramSize is the total DDR2 SDRAM size on the NuMaker-IIoT-NUC980G2: 128 MB.
//
//go:linkname ramSize runtime/goos.RamSize
var ramSize uint32 = 0x08000000 // 128 MB
