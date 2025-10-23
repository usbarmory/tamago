// NXP i.MX8MP initialization
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package imx8mp

import (
	"runtime"
	_ "unsafe"

	"github.com/usbarmory/tamago/dma"
)

// i.MX processor families
const (
	IMX8MP = 0x23
)

//go:linkname ramStackOffset runtime.ramStackOffset
var ramStackOffset uint32 = 0x100

var (
	// Processor family
	Family uint32

	// Flag native or emulated processor
	Native bool
)

// Init takes care of the lower level initialization triggered early in runtime
// setup (e.g. runtime.hwinit1).
func Init() {
	ramStart, _ := runtime.MemRegion()
	ARM64.Init(ramStart)

	// MMU initialization is required to take advantage of data cache
	ARM64.InitMMU()
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
	model := Model()

	switch model {
	case "i.MX8MP":
		OCOTP.Banks = 40
	}

	if !Native {
		return
	}
}
