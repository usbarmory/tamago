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

// UA0_BAUD_115200 provides the default BAUD divisor register values for a
// 115200 baud rate.
const UA0_BAUD_115200 = 0x30000066 // Mode:2 BRD:0x66 XIN:12MHz (baud = 12000000 / (0x66+2)

// UART0 is the primary console UART.
var UART0 = &uart.UART{
	Base: UA0_BA,
	Baud: UA0_BAUD_115200,
}
