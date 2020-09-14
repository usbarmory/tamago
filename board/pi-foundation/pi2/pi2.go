// Raspberry Pi 2 Support
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) the pi2 package authors
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package pi2

import (
	// Using go:linkname
	_ "unsafe"

	"github.com/f-secure-foundry/tamago/soc/bcm2835"

	// Ensure pi package is linked in, so client apps only *need* to
	// import this package
	_ "github.com/f-secure-foundry/tamago/board/pi-foundation"
)

const peripheralBase = 0x3f000000

// hwinit takes care of the lower level SoC initialization.
//go:linkname hwinit runtime.hwinit
func hwinit() {
	// Defer to generic BCM2835 initialization, with Pi 2
	// peripheral base address.
	bcm2835.Init(peripheralBase)
}
