// MCIMX6ULL-EVK support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkprintk
// +build !linkprintk

package mx6ullevk

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
)

// On the MCIMX6ULL-EVK the serial console is UART1, therefore standard
// output is redirected there.

//go:linkname printk runtime.printk
func printk(c byte) {
	imx6ul.UART1.Tx(c)
}
