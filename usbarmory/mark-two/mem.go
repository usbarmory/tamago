// USB armory Mk II support for tamago/arm
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build tamago,arm
// +build !linkramsize

// Go applications meant for tamago/arm on the USB armory Mk II simply need to
// import this package for all necessary hardware initialization.

package usbarmory

import (
	_ "unsafe"
)

// Applications can override ramSize with the `linkramsize` build tag. This is
// useful for applications that require large DMA descriptors and want to
// re-initialize tamago `mem` package in extermal RAM.

//go:linkname ramSize runtime.ramSize
var ramSize uint32 = 0x20000000 // 512 MB
