// QEMU sifive_u support for tamago/riscv64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkprintk

package sifive_u

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/soc/sifive/fu540"
)

//go:linkname printk runtime/goos.Printk
func printk(c byte) {
	fu540.UART0.Tx(c)
}
