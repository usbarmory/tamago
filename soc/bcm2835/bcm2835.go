// BCM2835 SoC support
// https://github.com/usbarmory/tamago
//
// Copyright (c) the bcm2835 package authors
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package bcm2835 provides support to Go bare metal unikernels written using
// the TamaGo framework on BCM2835/BCM2836 SoCs.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/usbarmory/tamago.
package bcm2835

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/arm"
)

// nanos - should be same value as arm/timer.go refFreq
const refFreq int64 = 1000000000

// DRAM_FLAG_NOCACHE disables caching by setting to high bits
const DRAM_FLAG_NOCACHE = 0xC0000000

// peripheralBase represents the (remapped) peripheral base address, it varies
// by model and it is therefore initialized (see Init) by individual board
// packages.
var peripheralBase uint32

// ARM processor instance
var ARM = &arm.CPU{}

//go:linkname ramStackOffset runtime.ramStackOffset
var ramStackOffset uint32 = 0x100000 // 1 MB

//go:linkname nanotime1 runtime.nanotime1
func nanotime1() int64 {
	return read_systimer()*ARM.TimerMultiplier + ARM.TimerOffset
}

// Init takes care of the lower level SoC initialization triggered early in
// runtime setup (e.g. runtime.hwinit).
func Init(base uint32) {
	peripheralBase = base

	ARM.Init()
	ARM.EnableVFP()

	// required when booting in SDP mode
	ARM.EnableSMP()

	// MMU initialization is required to take advantage of data cache
	ARM.InitMMU()
	ARM.EnableCache()

	ARM.TimerMultiplier = refFreq / SysTimerFreq
	ARM.TimerFn = read_systimer

	// initialize serial console
	MiniUART.Init()
}

// PeripheralAddress returns the absolute address for a peripheral. The Pi
// boards map 'bus addresses' to board specific base addresses but with
// consistent layout otherwise.
func PeripheralAddress(offset uint32) uint32 {
	return peripheralBase + offset
}
