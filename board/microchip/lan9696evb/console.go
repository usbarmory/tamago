// LAN969x 24-port EVB support for tamago/arm64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkprintk

package lan9696evb

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/soc/microchip/lan969x"
)

// On the LAN969x 24-port EVB FLEXCOM0 is connected as serial console through
// the on-board serial-to-USB (Type-C port) converter.  The standard output is
// redirected there.

//go:linkname printk runtime/goos.Printk
func printk(c byte) {
	lan969x.FLEXCOM0.Tx(c)
}
