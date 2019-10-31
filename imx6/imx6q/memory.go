// NXP i.MX6Q initialization
// https://github.com/inversepath/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build tamago,arm

package imx6q

import (
	_ "unsafe"
)

//go:linkname ramStart runtime.ramStart
var ramStart uint32 = 0x10000000

//go:linkname ramSize runtime.ramSize
var ramSize uint32 = 0x20000000 // FIXME: move to a board package

//go:linkname ramStackOffset runtime.ramStackOffset
var ramStackOffset uint32 = 0x100000
