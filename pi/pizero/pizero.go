// Raspberry Pi Zero Support
// https://github.com/f-secure-foundry/tamago
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package pizero

import (
	// Using go:linkname
	_ "unsafe"

	"github.com/f-secure-foundry/tamago/bcm2835"

	// Ensure pi package is linked in, so client apps only *need* to
	// import this package
	_ "github.com/f-secure-foundry/tamago/pi"
)

const peripheralBase uint32 = 0x20000000

// hwinit takes care of the lower level SoC initialization.
//
// Triggered early in runtime setup, care must be taken to ensure that
// no heap allocation is performed (e.g. defer is not possible).
//go:linkname hwinit runtime.hwinit
func hwinit() {
	// Defer to generic BCM2835 initialization, with Pi Zero
	// peripheral base address.
	bcm2835.HardwareInit(peripheralBase)
}
