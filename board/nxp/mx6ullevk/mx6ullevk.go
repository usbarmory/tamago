// MCIMX6ULL-EVK support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package mx6ullevk provides hardware initialization, automatically on import,
// for the NXP MCIMX6ULL-EVK evaluation board.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/usbarmory/tamago.
package mx6ullevk

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/soc/imx6"
	"github.com/usbarmory/tamago/soc/imx6/imx6ul"
)

// Peripheral instances
var (
	I2C1 = imx6ul.I2C1
	I2C2 = imx6ul.I2C2

	UART1 = imx6ul.UART1
	UART2 = imx6ul.UART2

	USB1 = imx6ul.USB1
	USB2 = imx6ul.USB2

	// SD1 is the base board full size SD instance
	USDHC1 = imx6ul.USDHC1
	USDHC2 = imx6ul.USDHC2
)

// Init takes care of the lower level SoC initialization triggered early in
// runtime setup.
//
//go:linkname Init runtime.hwinit
func Init() {
	imx6.Init()

	// initialize console
	imx6ul.UART1.Init()
}
