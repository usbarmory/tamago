// Raspberry Pi support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) the pi package authors
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkprintk
// +build !linkprintk

package pi

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/soc/bcm2835"
)

//go:linkname printk runtime.printk
func printk(c byte) {
	bcm2835.MiniUART.Tx(c)
}
