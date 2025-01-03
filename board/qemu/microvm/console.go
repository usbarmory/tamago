// QEMU microvm support for tamago/amd64
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkprintk

package microvm

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/amd64"
)

const COM1 = 0x3f8

//go:linkname printk runtime.printk
func printk(c byte) {
	amd64.Out(COM1, uint32(c))
}
