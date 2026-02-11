// Google Compute Engine support for tamago/amd64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkprintk

package gcp

import (
	_ "unsafe"
)

//go:linkname printk runtime/goos.Printk
func printk(c byte) {
	UART0.Tx(c)
}
