// 8MPLUSLPD4-EVK support for tamago/arm64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkprintk

package imx8mpevk

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/soc/nxp/imx8mp"
)

// On the 8MPLUSLPD4-EVK the serial console is UART1, therefore standard
// output is redirected there.

//go:linkname printk runtime/goos.Printk
func printk(c byte) {
	imx8mp.UART1.Tx(c)
}
