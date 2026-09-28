// QEMU microvm support for tamago/amd64
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

//go:linkname ramSize runtime/goos.RamSize
var ramSize uint64 = 0xb0000000 - dmaSize // 2560 MiB

// On amd64 memory above 4 GiB can be accessed by relocating the heap memory
// start address, the following example enables 8 GiB memory allocation in
// extended memory and must be hooked in goos.Hwinit0:
//
//   goos.RamStart = 0x1_0000_0000
//   goos.RamSize  = 8 << 30
//   goos.Bloc     = uintptr(goos.RamStart)
//   goos.BlocMax  = uintptr(goos.RamStart + goos.RamSize)
//
// Note that this does not cover any MMU re-configuration that might be
// required depending on the boot state of PDPT entries (see amd64/init.s).
