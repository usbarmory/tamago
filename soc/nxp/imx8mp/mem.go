// NXP i.MX8MP initialization
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkramstart

package imx8mp

import (
	_ "unsafe"
)

//go:linkname ramStart runtime.ramStart
var ramStart uint32 = DDR_BASE
