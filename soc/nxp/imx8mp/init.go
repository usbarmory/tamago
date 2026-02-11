// NXP i.MX8MP initialization
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package imx8mp

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/dma"
)

// i.MX8 processor families
const (
	// i.MX 8M Mini
	IMX8MMD = 0x8201
	IMX8MMQ = 0x8241

	// i.MX 8M Plus
	IMX8MPD = 0x8203
	IMX8MPQ = 0x8240
)

//go:linkname ramStackOffset runtime/goos.RamStackOffset
var ramStackOffset uint64 = 0x100

var (
	// Processor family
	Family uint32

	// Flag native or emulated processor
	Native bool
)

// Init takes care of the lower level initialization triggered early in runtime
// setup (e.g. runtime/goos.Hwinit1).
func Init() {
	ARM64.Init()
	ARM64.EnableCache()

	initTimers()
}

func init() {
	// Initialize watchdogs, this must be done within 16 seconds to clear
	// their power-down counter event
	// (p4085, 6.6.2.4 Power-down counter event, IMX8MPRM).
	WDOG1.Init()
	WDOG2.Init()
	WDOG3.Init()

	// use internal OCRAM (iRAM) as default DMA region
	dma.Init(OCRAM_START, OCRAM_SIZE)

	OCOTP.Init()

	switch Family {
	case IMX8MMD, IMX8MMQ:
		OCOTP.Banks = 14
	case IMX8MPD, IMX8MPQ:
		OCOTP.Banks = 40
	}
}
