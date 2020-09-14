// Raspberry Pi 2 Support
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) the pi2 package authors
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build !linkramsize

package pi2

import (
	// Using go:linkname
	_ "unsafe"
)

//go:linkname ramSize runtime.ramSize
var ramSize uint32 = 0x40000000 - 0x4C00000 // 1GB - 76MB (VideoCore)
