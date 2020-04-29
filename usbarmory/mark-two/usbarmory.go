// USB armory Mk II support for tamago/arm
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build tamago,arm

// Go applications meant for tamago/arm on the USB armory Mk II simply need to
// import this package for all necessary hardware initialization.

package usbarmory

import (
	_ "unsafe"

	"github.com/f-secure-foundry/tamago/imx6"
	"github.com/f-secure-foundry/tamago/imx6/usdhc"
	_ "github.com/f-secure-foundry/tamago/imx6/imx6ul"
)

//go:linkname ramSize runtime.ramSize
var ramSize uint32 = 0x20000000 // 512 MB

// external uSD
var SD = usdhc.USDHC1
// internal eMMC
var MMC = usdhc.USDHC2

func init() {
	SD.Init(4, false)
	MMC.Init(8, true)
}

//go:linkname printk runtime.printk
func printk(c byte) {
	imx6.UART2.Write(c)
}
