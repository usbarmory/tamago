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
	// Emit CR before LF so output is legible on serial terminals that do
	// not perform implicit CR insertion (raw mode, minicom, screen, etc.).
	if c == '\n' {
		nuc980.UART0.Tx('\r')
	}
	nuc980.UART0.Tx(c)
}
