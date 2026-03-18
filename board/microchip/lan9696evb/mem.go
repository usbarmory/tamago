// LAN969x 24-port EVB support for tamago/arm64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkramsize

package lan9696evb

import (
	_ "unsafe"
)

// Applications can override ramSize with the `linkramsize` build tag.
//
// This is useful when large DMA descriptors are required to re-initialize
// tamago `dma` package in external RAM.

// The LAN969x 24-port EVB features 1GB DDR4 RAM.

// Align with mem-size in:
//
//	https://github.com/microchip-ung/arm-trusted-firmware/tree/main/plat/microchip/lan969x/fdts/lan969x-ddr.dtsi
//
//go:linkname ramSize runtime/goos.RamSize
var ramSize uint64 = 0x38000000 // 896MB
