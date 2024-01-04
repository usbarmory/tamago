// USB armory Mk II support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package mk2 provides hardware initialization, automatically on import, for
// the USB armory Mk II single board computer.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/usbarmory/tamago.
package mk2

import (
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"

	_ "unsafe"
)

// USB armory Mk II model revisions, WithSecure burns model information in the
// MSB of OTP fuses bank 4 word 2 (OCOTP_MAC0), returned by Model().
const (
	// USB armory Mk II (rev. β) - UA-MKII-β
	BETA = iota
	// USB armory Mk II (rev. γ) - UA-MKII-γ
	GAMMA
	// USB armory Mk II LAN - UA-MKII-LAN
	LAN
)

// Peripheral instances
var (
	ENET2 = imx6ul.ENET2

	I2C1 = imx6ul.I2C1
	I2C2 = imx6ul.I2C2

	UART1 = imx6ul.UART1
	UART2 = imx6ul.UART2

	USB1 = imx6ul.USB1
	USB2 = imx6ul.USB2

	USDHC1 = imx6ul.USDHC1
	USDHC2 = imx6ul.USDHC2
)

func model() uint32 {
	if !imx6ul.Native {
		return 0
	}

	// WithSecure burns model information in the MSB of OTP fuses bank 4
	// word 2 (OCOTP_MAC0).
	mac0, _ := imx6ul.OCOTP.Read(4, 2)
	return mac0 >> 24
}

// Model returns the USB armory model name, to further detect SoC variants
// imx6ul.Model() can be used.
func Model() (m uint32, s string) {
	m = model()

	switch m {
	case BETA:
		s = "UA-MKII-β"
	case GAMMA:
		s = "UA-MKII-γ"
	case LAN:
		s = "UA-MKII-LAN"
	default:
		s = "unknown"
	}

	return
}

// Init takes care of the lower level SoC initialization triggered early in
// runtime setup.
//
//go:linkname Init runtime.hwinit
func Init() {
	// initialize SoC
	imx6ul.Init()

	// initialize serial console
	imx6ul.UART2.Init()
}
