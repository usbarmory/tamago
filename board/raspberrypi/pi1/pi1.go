// Raspberry Pi 1 support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) the pi1 package authors
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package pi1 provides hardware initialization, automatically on import,
// for the Raspberry Pi 1 single board computer.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/usbarmory/tamago.
package pi1

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/board/raspberrypi"
	"github.com/usbarmory/tamago/soc/bcm2835"
)

const peripheralBase = 0x20000000

type board struct{}

// Board provides access to the capabilities of the Pi 1.
var Board pi.Board = &board{}

// Init takes care of the lower level SoC initialization triggered early in
// runtime setup.
//
//go:linkname Init runtime.hwinit
func Init() {
	// Defer to generic BCM2835 initialization, with Pi 1
	// peripheral base address.
	bcm2835.Init(peripheralBase)
}
