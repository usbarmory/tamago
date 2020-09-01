// USB armory Mk II support for tamago/arm
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package usbarmory provides hardware initialization, automatically on import,
// for the USB armory Mk II single board computer.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/f-secure-foundry/tamago.
package usbarmory

import (
	_ "unsafe"

	"github.com/f-secure-foundry/tamago/soc/imx6"
	_ "github.com/f-secure-foundry/tamago/soc/imx6/imx6ul"
)

// Model returns the USB armory model name, to further detect SoC variants
// imx6.Model() can be used.
func Model() (model string) {
	// for now only β exists in the wild
	return "UA-MKII-β"
}

// Init takes care of the lower level SoC initialization triggered early in
// runtime setup, care must be taken to ensure that no heap allocation is
// performed (e.g. defer is not possible).
//
//go:linkname Init runtime.hwinit
func Init() {
	imx6.Init()

	// initialize serial console
	imx6.UART2.Init()
}
