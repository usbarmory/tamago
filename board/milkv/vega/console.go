// MilkV Vega support for tamago/riscv64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkprintk

package vega

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/soc/fisilink/fsl91030"
)

//go:linkname printk runtime/goos.Printk
func printk(c byte) {
	fsl91030.UART0.Tx(c)
}
