// USB armory Mk II support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkprintk
// +build !linkprintk

package mk2

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
)

// On the USB armory Mk II the serial console is UART2, therefore standard
// output is redirected there.
//
// On models UA-MKII-β and UA-MKII-γ the console is exposed through the USB
// Type-C receptacle and available only in debug accessory mode (see
// EnableDebugAccessory()).
//
// On model UA-MKII-LAN the console is exposed through test pads.

//go:linkname printk runtime.printk
func printk(c byte) {
	imx6ul.UART2.Tx(c)
}
