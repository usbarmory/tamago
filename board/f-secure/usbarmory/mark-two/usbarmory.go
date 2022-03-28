// USB armory Mk II support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package usbarmory provides hardware initialization, automatically on import,
// for the USB armory Mk II single board computer.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/usbarmory/tamago.
package usbarmory

import (
	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/soc/imx6"
	_ "github.com/usbarmory/tamago/soc/imx6/imx6ul"

	_ "unsafe"
)

const OCOTP_MAC0 = 0x021bc620

const (
	REV_BETA = iota
	REV_GAMMA
)

// Model returns the USB armory model name, to further detect SoC variants
// imx6.Model() can be used.
func Model() (model string) {
	// F-Secure burns model information in the MSB of OTP fuses bank 4 word 2.
	mac0 := reg.Read(OCOTP_MAC0)

	switch mac0 >> 24 {
	case REV_GAMMA:
		return "UA-MKII-γ"
	default:
		return "UA-MKII-β"
	}
}

// Init takes care of the lower level SoC initialization triggered early in
// runtime setup.
//
//go:linkname Init runtime.hwinit
func Init() {
	// initialize SoC
	imx6.Init()

	// initialize serial console
	imx6.UART2.Init()
}
