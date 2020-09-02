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

const (
	// nanos - should be same value as arm/timer.go refFreq
	refFreq int64 = 1000000000
)

// peripheralBase is the (remapped) peripheral base address.
//
// In Raspberry Pi, the VideoCore chip is responsible for
// bootstrapping.  In Pi2+, it remaps registers from their
// hardware 'bus' address to the 0x3f000000 'physical'
// address.  In Pi Zero, registers start at 0x20000000.
//
// This varies by model, hence variable so can be overridden
// at runtime.
//go:linkname peripheralBase runtime.peripheralBase
var peripheralBase uint32

// ARM processor instance
var ARM = &arm.CPU{}

//go:linkname nanotime1 runtime.nanotime1
func nanotime1() int64 {
	return int64(read_systimer() * ARM.TimerMultiplier)
}

// Init takes care of the lower level SoC initialization.
//
// Triggered early in runtime setup, care must be taken to ensure that
// no heap allocation is performed (e.g. defer is not possible).
func Init(baseAddress uint32) {

	// The peripheral base address differs by board
	peripheralBase = baseAddress

	ARM.Init()
	ARM.EnableVFP()

	// required when booting in SDP mode
	ARM.EnableSMP()

	ARM.CacheEnable()

	ARM.TimerMultiplier = refFreq / SysTimerFreq
	ARM.TimerFn = read_systimer

	uartInit()
}

// PeripheralAddress gets the absolute address for a peripheral.
//
// The Pi boards map 'bus addresses' to different memory addresses
// by board, but have a consistent layout otherwise.
//
func PeripheralAddress(offset uint32) uint32 {
	return peripheralBase + offset
}
