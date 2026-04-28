// Nuvoton NUC980 UART driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// UART0 console driver for the NUC980 SoC.
//
// Pin mux: UART0_RXD = GPF11, UART0_TXD = GPF12 (MFP function 1).
// See NUC980 Series Datasheet, p. 130 (§ 6.2.7, SYS_GPF_MFPH register) and
// p. 190 (§ 6.14 UART Interface Controller).
// Baud rate: 115200, 8N1, configured for XIN = 12 MHz input.

package nuc980

import (
	"github.com/usbarmory/tamago/internal/reg"
)

// System Manager / Pin Mux registers
//
// NUC980 Series Datasheet, p. 93 (§ 6.2.6 register map).
const (
	SYS_BA = 0xB0000000

	// GPF multi-function pin register (high: GPF8..GPF15).
	// Each pin occupies 4 bits; GPF8=[3:0], GPF9=[7:4], GPF10=[11:8],
	// GPF11=[15:12], GPF12=[19:16], GPF13=[23:20], GPF14=[27:24], GPF15=[31:28].
	// [15:12] = GPF11 function select (1 = UART0_RXD)
	// [19:16] = GPF12 function select (1 = UART0_TXD)
	// NUC980 Series Datasheet, p. 130 (§ 6.2.7 SYS_GPF_MFPH).
	REG_MFP_GPF_H = SYS_BA + 0x09C
)

// UART0 register offsets from base 0xB0070000
//
// NUC980 Series Datasheet, p. 190 (§ 6.14 UART Interface Controller).
const (
	UA0_BA = 0xB0070000

	UA_RBR  = UA0_BA + 0x00 // Receive Buffer Register
	UA_THR  = UA0_BA + 0x00 // Transmit Holding Register (same offset)
	UA_IER  = UA0_BA + 0x04 // Interrupt Enable Register
	UA_FCR  = UA0_BA + 0x08 // FIFO Control Register
	UA_LCR  = UA0_BA + 0x0C // Line Control Register
	UA_MCR  = UA0_BA + 0x10 // Modem Control Register
	UA_FSR  = UA0_BA + 0x18 // FIFO Status Register
	UA_BAUD = UA0_BA + 0x24 // Baud Rate Divisor Register
)

// BAUD register: Mode 2 (bits [31:30] = 0b11), BRD = 0x66 = 102.
// XIN = 12 MHz → baud = 12000000 / (102+2) = 115384 ≈ 115200.
// NUC980 Series Datasheet, p. 190 (§ 6.14, UART_BAUD register).
const UA_BAUD_115200 = 0x30000066

// UA_FSR TX_FULL bit: set when Tx FIFO is full.
// NUC980 Series Datasheet, p. 190 (§ 6.14, UART_FSR register, bit 23).
const UA_FSR_TX_FULL = 1 << 23

// UART represents a NUC980 UART instance.
type UART struct {
	thr uint32
	fsr uint32
}

// UART0 is the primary console UART.
var UART0 = &UART{
	thr: UA_THR,
	fsr: UA_FSR,
}

// Init configures UART0 for 115200 8N1 operation using XIN (12 MHz).
//
// Pin mux for GPF11/GPF12 and the UART0 APB clock are configured
// from assembly in the board cpuinit.
func (hw *UART) Init() {
	// 8-bit word length, no parity, 1 stop bit.
	reg.Write(UA_LCR, 0x3)

	// Reset and enable TX/RX FIFOs.
	reg.Write(UA_FCR, 0x6)

	// Set baud rate: 115200 from 12 MHz XIN.
	reg.Write(UA_BAUD, UA_BAUD_115200)
}

// Tx transmits a single byte on UART0, blocking until the TX FIFO has room.
func (hw *UART) Tx(c byte) {
	for reg.Read(hw.fsr)&UA_FSR_TX_FULL != 0 {
	}
	reg.Write(hw.thr, uint32(c))
}

// Write transmits buf to UART0.
func (hw *UART) Write(buf []byte) {
	for _, c := range buf {
		hw.Tx(c)
	}
}
