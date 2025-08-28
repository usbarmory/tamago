// Firecracker microvm support for tamago/amd64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkramsize

package microvm

import (
	_ "unsafe"
)

// Applications can override ramSize with the `linkramsize` build tag.
//
// This is useful when large DMA descriptors are required to re-initialize
// tamago `dma` package in external RAM.

//go:linkname ramSize runtime.ramSize
var ramSize uint64 = 0x40000000 // 1GB
