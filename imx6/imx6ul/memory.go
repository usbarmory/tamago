// NXP i.MX6UL/i.MX6ULL initialization
// https://github.com/inversepath/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build tamago,arm

package imx6ul

import (
	_ "unsafe"
)

//go:linkname ramStart runtime.ramStart
var ramStart uint32 = 0x80000000

// ramSize defined in board package
//go:linkname ramStackOffset runtime.ramStackOffset
var ramStackOffset uint32 = 0x100
