// NXP i.MX6UL/i.MX6ULL/i.MX6Q support
// https://github.com/inversepath/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build tamago,arm

package imx6

import (
	"unsafe"
)

const USB_ANALOG_DIGPROG uint32 = 0x020c8260

const IMX6Q = 0x63
const IMX6UL = 0x64
const IMX6ULL = 0x65

var Family uint32
var Native bool

// hwinit takes care of the lower level SoC initialization triggered early in
// runtime setup, care must be taken to ensure that no heap allocation is
// performed (e.g. defer is not possible).
//go:linkname hwinit runtime.hwinit
func hwinit() {
	_, fam, revMajor, revMinor := SiliconVersion()
	Family = fam

	if revMajor != 0 || revMinor != 0 {
		Native = true
	}

	switch Family {
	case IMX6Q:
		initGlobalTimers()
	case IMX6UL, IMX6ULL:
		initGenericTimers()
	default:
		initGlobalTimers()
	}

	enableVFP()

	return
}

//go:linkname initRNG runtime.initRNG
func initRNG() {
	if Family == IMX6ULL && Native {
		RNGB.Init()
		getRandomDataFn = RNGB.getRandomData
	} else {
		getRandomDataFn = getLCGData
	}
}

// SiliconVersion returns the SoC silicon version information
// (p3945, 57.4.11 Chip Silicon Version (USB_ANALOG_DIGPROG), IMX6ULLRM).
func SiliconVersion() (sv, family, revMajor, revMinor uint32) {
	sv = (*(*uint32)(unsafe.Pointer(uintptr(USB_ANALOG_DIGPROG))))

	family = (sv >> 16) & 0xff
	revMajor = (sv >> 8) & 0xff
	revMinor = sv & 0xff

	return
}

// Model returns the SoC model name.
func Model() (model string) {
	switch Family {
	case IMX6Q:
		model = "i.MX6Q"
	case IMX6UL:
		model = "i.MX6UL"
	case IMX6ULL:
		model = "i.MX6ULL"
	default:
		model = "unknown"
	}

	return
}
