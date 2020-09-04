// Raspberry Pi Zero Support
// https://github.com/f-secure-foundry/tamago
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package pizero

import (
	// use of go:linkname
	_ "unsafe"
)

//go:linkname ramStart runtime.ramStart
var ramStart uint32 = 0x0

//go:linkname ramStackOffset runtime.ramStackOffset
var ramStackOffset uint32 = 0x100000 // 1 MB

//go:linkname ramSize runtime.ramSize
var ramSize uint32 = 0x20000000 - 0x04000000 // 512 MB - 64MB GPU (VideoCore)
