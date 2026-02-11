// MCIMX6ULL-EVK support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkramsize

package mx6ullevk

import (
	_ "unsafe"
)

// Applications can override ramSize with the `linkramsize` build tag.
//
// This is useful when large DMA descriptors are required to re-initialize
// tamago `dma` package in external RAM.

// The MCIMX6ULL-EVK features a single 512MB DDR3 RAM module.

//go:linkname ramSize runtime/goos.RamSize
var ramSize uint32 = 0x20000000 // 512MB
