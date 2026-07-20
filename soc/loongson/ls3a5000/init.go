// Loongson 3A5000/LS7A support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package ls3a5000

import (
	_ "unsafe"
)

//go:linkname ramStackOffset runtime/goos.RamStackOffset
var ramStackOffset uint64 = 0x100
