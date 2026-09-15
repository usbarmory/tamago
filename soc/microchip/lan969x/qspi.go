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

	CPU_QSPI0_GCK = CPU_BASE + 0xb4
)

// HSIO pad configuration
const (
	HSIO_QSPI0_PAD_CONFIG          = 0x28
	HSIO_QSPI0_PAD_REGISTER_STRIDE = 0x4
)

const (
	qspi0PadCount    = 6
	qspi0ParentClock = 1_000_000_000
	qspi0TargetClock = 100_000_000
)

// ConfigureQSPI0Pins selects the dedicated nCS, SCK, and IO0 through IO3 pins
// and assigns their shared serial interface to QSPI0. Call it before QSPI0.Init
// and only while the controller is idle.
func ConfigureQSPI0Pins() {
	for i := range qspi0PadCount {
		reg.Write(HSIO_BASE+uint32(i)*HSIO_QSPI0_PAD_REGISTER_STRIDE, HSIO_QSPI0_PAD_CONFIG)
	}

	reg.Set(CPU_GENERAL_CTRL, GENERAL_CTRL_IF_SI_OWNER)
}
