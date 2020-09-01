// USB armory Mk II support for tamago/arm
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package usbarmory

import (
	_ "unsafe"

	"github.com/f-secure-foundry/tamago/imx6"
)

// On the USB armory Mk II the serial console is UART2, therefore standard
// output is redirected there.

//go:linkname printk runtime.printk
func printk(c byte) {
	imx6.UART2.Tx(c)
}
