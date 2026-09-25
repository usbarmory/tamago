// Microchip LAN969x configuration and support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package lan969x

import "github.com/usbarmory/tamago/internal/reg"

// CPU registers
const (
	CPU_GENERAL_CTRL         = CPU_BASE + 0x8c
	GENERAL_CTRL_IF_SI_OWNER = 0
)

// HSIOWRAP registers
const (
	HSIOWRAP_GPIO_CFG = 0x00
	GPIO_CFG_ST       = 5
	GPIO_CFG_PU       = 3
	GPIO_CFG_DS       = 0
)

const (
	qspi0ClockID     = 0
	qspi0ParentClock = 1_000_000_000
	qspi0TargetClock = 100_000_000
	qspi0MMAPSize    = 0x08000000

	// dedicated nCS, SCK, and IO0 through IO3 pads
	qspi0PadCount = 6
	// minimum drive strength, pull-up, Schmitt trigger
	qspi0PadConfig = 0<<GPIO_CFG_DS | 1<<GPIO_CFG_PU | 1<<GPIO_CFG_ST
)

// ConfigureQSPI0Pins sets the electrical configuration of the dedicated nCS,
// SCK, and IO0 through IO3 pads and assigns their shared serial interface to
// QSPI0. Call it before QSPI0.Init and only while the controller is idle.
func ConfigureQSPI0Pins() {
	for i := range uint32(qspi0PadCount) {
		reg.Write(HSIO_BASE+HSIOWRAP_GPIO_CFG+i*4, qspi0PadConfig)
	}

	reg.Set(CPU_GENERAL_CTRL, GENERAL_CTRL_IF_SI_OWNER)
}
