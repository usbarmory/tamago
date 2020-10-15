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

	"github.com/f-secure-foundry/tamago/soc/imx6"
)

const (
	// I2C address of FUSB303
	FUSB303_ADDR = 0x31

	FUSB303_CONTROL1 = 0x05
	CONTROL1_ENABLE  = 3
)

// On the USB armory Mk II the serial console is UART2, therefore standard
// output is redirected there.

//go:linkname printk runtime.printk
func printk(c byte) {
	imx6.UART2.Tx(c)
}

// EnableDebugAccessory enables the debug detect on FUSB303,
// this allows to have console debug when using Debug accessory.
func EnableDebugAccessory() (err error) {
	a, err := imx6.I2C1.Read(FUSB303_ADDR, FUSB303_CONTROL1, 1, 1)

	if err != nil {
		return
	}

	a[0] |= 1 << CONTROL1_ENABLE
	err = imx6.I2C1.Write(a, FUSB303_ADDR, FUSB303_CONTROL1, 1)

	return
}
