// Banana Pi BPI-F3 support for tamago/riscv64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkramsize

package f3

import (
	_ "unsafe"
)

// The BPI-F3 is available with 2, 4, 8 or 16 GB of LPDDR4x DRAM, the default
// size below is therefore conservative and covers all variants. Applications
// can override ramSize with the linkramsize build tag.

//go:linkname ramSize runtime/goos.RamSize
var ramSize uint64 = 0x20000000 // 512 MB
