// x86-64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package amd64

import (
	_ "unsafe"
)

// defined in timer.s
func read_tsc() int64

//go:linkname nanotime1 runtime.nanotime1
func nanotime1() int64 {
	return read_tsc()
}

