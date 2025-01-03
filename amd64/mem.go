// x86-64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkramstart

package amd64

import (
	_ "unsafe"
)

//go:linkname ramStart runtime.ramStart
var ramStart uint64 = 0x10000000
