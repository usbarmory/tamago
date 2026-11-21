// SpacemiT K1 memory layout
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkramstart

package k1

import (
	_ "unsafe"
)

// K1 DRAM is mapped at 0x00000000 (p13, Section 2.2.2 DDR, K1 Datasheet), the first 32 MB are conventionally occupied
// by the boot ROM loaded first stage (FSBL, OpenSBI, U-Boot)
// and are therefore left outside of the runtime managed memory.
// Applications can override ramStart with the linkramstart build tag.

//go:linkname ramStart runtime/goos.RamStart
var ramStart uint64 = 0x02000000
