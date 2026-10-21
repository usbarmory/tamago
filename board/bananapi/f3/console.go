// Banana Pi BPI-F3 support for tamago/riscv64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkprintk

package f3

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/soc/spacemit/k1"
)

//go:linkname printk runtime/goos.Printk
func printk(c byte) {
	k1.UART0.TrySyncTx(c, 1<<22)

	// The runtime terminates lines with a line feed alone, which advances
	// a terminal to the next row without returning to the first column,
	// staircasing the output.
	if c == 0x0a { // LF
		k1.UART0.TrySyncTx(0x0d, 1<<22) // CR
	}
}
