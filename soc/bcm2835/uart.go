// BCM2835 mini-UART driver
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) the bcm2835 package authors
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// +build !linkprintk

package bcm2835

import (
	_ "unsafe"
)

//go:linkname printk runtime.printk
func printk(c byte) {
	MiniUART.Tx(c)
}
