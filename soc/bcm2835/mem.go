// BCM2835 SoC support
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) the bcm2835 package authors
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// +build !linkramstart

package bcm2835

import (
	_ "unsafe"
)

//go:linkname ramStart runtime.ramStart
var ramStart uint32 = 0x00100000
