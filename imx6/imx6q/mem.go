// NXP i.MX6Q initialization
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build !linkramstart

// Package imx6q provides hardware initialization, automatically on import, for
// the i.MX6Q family of System-on-Chip components.

package imx6q

import (
	_ "unsafe"
)

//go:linkname ramStart runtime.ramStart
var ramStart uint32 = 0x10000000
