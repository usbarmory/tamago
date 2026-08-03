// QEMU LoongArch virt support for tamago/loong64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkprintk

package virt

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/soc/loongson/ls3a5000"
)

//go:linkname printk runtime/goos.Printk
func printk(c byte) {
	ls3a5000.UART0.Tx(c)
}
