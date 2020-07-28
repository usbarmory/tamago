// BCM2835 SOC support
// https://github.com/f-secure-foundry/tamago
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package bcm2835

import (
	// using go:linkname
	_ "unsafe"

	"github.com/f-secure-foundry/tamago/arm"
)

// PeripheralBase is the (remapped) peripheral base address.
//
// In Raspberry Pi, the VideoCore chip is responsible for
// bootstrapping.  In Pi2+, it remaps registers from their
// hardware 'bus' address to the 0x3f000000 'physical'
// address.  In Pi Zero, registers start at 0x20000000.
//
// This varies by model, hence variable so can be overridden
// at runtime.
//go:linkname PeripheralBase runtime.PeripheralBase
var PeripheralBase uint32

// ARM processor instance
var ARM = &arm.CPU{}

//go:linkname nanotime1 runtime.nanotime1
func nanotime1() int64 {
	return int64(read_systimer() * ARM.TimerMultiplier)
}

// HardwareInit takes care of the lower level SoC initialization.
//
// Triggered early in runtime setup, care must be taken to ensure that
// no heap allocation is performed (e.g. defer is not possible).
func HardwareInit(peripheralBase uint32) {

	// The peripheral base address differs by board
	PeripheralBase = peripheralBase

	ARM.Init()
	ARM.EnableVFP()

	// required when booting in SDP mode
	ARM.EnableSMP()

	ARM.CacheEnable()

	ARM.InitSpecificTimer(read_systimer, SysTimerFreq)

	uartInit()
}
