// NXP i.MX6UL initialization
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkramstart
// +build !linkramstart

package imx6ul

import (
	_ "unsafe"
)

//go:linkname ramStart runtime.ramStart
var ramStart uint32 = MMDC_BASE
