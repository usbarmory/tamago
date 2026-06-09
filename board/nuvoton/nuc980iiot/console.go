// NuMaker-IIoT-NUC980G2 board support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkprintk

package nuc980iiot

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/soc/nuvoton/nuc980"
)

//go:linkname printk runtime/goos.Printk
func printk(c byte) {
	nuc980.UART0.Tx(c)
}
