// Raspberry Pi 2 support for tamago/arm
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) the pi2 package authors
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package pi2 provides hardware initialization, automatically on import, for
// the Raspberry Pi 2 single board computer.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/f-secure-foundry/tamago.
package pi2

import (
	// Using go:linkname
	_ "unsafe"

	"github.com/f-secure-foundry/tamago/soc/bcm2835"

	// Ensure pi package is linked in, so client apps only *need* to
	// import this package
	"github.com/f-secure-foundry/tamago/board/pi-foundation"
)

const peripheralBase = 0x3f000000

type board struct{}

// Board provides access to the capabilities of the Pi2.
var Board pi.Board = &board{}

// hwinit takes care of the lower level SoC initialization.
//go:linkname hwinit runtime.hwinit
func hwinit() {
	// Defer to generic BCM2835 initialization, with Pi 2
	// peripheral base address.
	bcm2835.Init(peripheralBase)
}
