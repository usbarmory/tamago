// USB armory Mk II support for tamago/arm
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// +build !linkramsize

package usbarmory

import (
	_ "unsafe"
)

// Applications can override ramSize with the `linkramsize` build tag.
//
// This is useful when large DMA descriptors are required to re-initialize
// tamago `mem` package in external RAM.

// The standard USB armory Mk II features a single 512MB DDR3 RAM module.

//go:linkname ramSize runtime.ramSize
var ramSize uint32 = 0x20000000 // 512 MB
