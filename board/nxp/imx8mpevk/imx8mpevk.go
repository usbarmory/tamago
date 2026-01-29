// 8MPLUSLPD4-EVK support for tamago/arm64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package imx8mpevk provides hardware initialization, automatically on import,
// for the NXP 8MPLUSLPD4-EVK evaluation board.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm64` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package imx8mpevk

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/soc/nxp/imx8mp"
)

// Peripheral instances
var (
	ENET1 = imx8mp.ENET1
	UART1 = imx8mp.UART1
)

// Init takes care of the lower level initialization triggered early in runtime
// setup (post World start).
//
//go:linkname Init runtime.hwinit1
func Init() {
	imx8mp.Init()

	// initialize console
	imx8mp.UART1.Init()
}
