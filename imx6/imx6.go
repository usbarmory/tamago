// NXP i.MX6UL/i.MX6ULL/i.MX6Q support
// https://github.com/f-secure-foundry/tamago
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
	"encoding/binary"
	"unsafe"

	"github.com/f-secure-foundry/tamago/imx6/internal/reg"
)

const USB_ANALOG_DIGPROG uint32 = 0x020c8260
const WDOG1 uint32 = 0x020bc000
const OCOTP_CFG0 = 0x021bc410
const OCOTP_CFG1 = 0x021bc420

const IMX6Q = 0x63
const IMX6UL = 0x64
const IMX6ULL = 0x65

var Family uint32
var Native bool

var features processorFeatures

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

	features.read()
	UART2.init(UART2_BASE, UART_DEFAULT_BAUDRATE)

	switch Family {
	case IMX6Q:
		initGlobalTimers()
	case IMX6UL, IMX6ULL:
		if (Native) {
			// U-Boot value for i.MX6 family (8.0MHz)
			initGenericTimers(8000000)
		} else {
			initGenericTimers(0)
		}
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
	sv = reg.Read(USB_ANALOG_DIGPROG)

	family = (sv >> 16) & 0xff
	revMajor = (sv >> 8) & 0xff
	revMinor = sv & 0xff

	return
}

// UniqueID returns the NXP SoC Device Unique 64-bit ID
func UniqueID() (uid [8]byte) {
	cfg0 := (*uint32)(unsafe.Pointer(uintptr(OCOTP_CFG0)))
	cfg1 := (*uint32)(unsafe.Pointer(uintptr(OCOTP_CFG1)))

	binary.LittleEndian.PutUint32(uid[0:4], *cfg0)
	binary.LittleEndian.PutUint32(uid[4:8], *cfg1)

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

// Reboot resets the watchdog timer causing the SoC to restart.
func Reboot() {
	reg := (*uint16)(unsafe.Pointer(uintptr(WDOG1)))
	*reg = 0x0000
}
