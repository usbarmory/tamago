// Cloud Hypervisor support for tamago/amd64
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkprintk

package vm

import (
	_ "unsafe"
)

//go:linkname printk runtime.printk
func printk(c byte) {
	UART0.Tx(c)

	if c == 0x0a { // LF
		UART0.Tx(0x0d) // CR
	}
}
