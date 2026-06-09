// Nuvoton NUC980 UART support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package nuc980

import (
	"github.com/usbarmory/tamago/soc/nuvoton/uart"
)

// UART0 register base.
const UA0_BA = 0xb0070000

// UA0_BAUD_115200 selects BAUD Mode 2 (bits [31:30] = 0b11) with BRD = 0x66.
// With XIN = 12 MHz: baud = 12000000 / (0x66+2) = 115384 ≈ 115200.
const UA0_BAUD_115200 = 0x30000066

// UART0 is the primary console UART.
//
// Pin mux (UART0_RXD = GPF11, UART0_TXD = GPF12, MFP function 1) and the
// UART0 APB clock are configured from assembly in the board cpuinit.
var UART0 = &uart.UART{
	Base: UA0_BA,
	Baud: UA0_BAUD_115200,
}
