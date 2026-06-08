// Nuvoton UART driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package uart implements a driver for the UART Interface Controller found on
// Nuvoton SoCs adopting the following reference specifications:
//   - NUC980 Series Datasheet - Rev 1.24
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package uart

import (
	"github.com/usbarmory/tamago/internal/reg"
)

// UART register offsets (from UART.Base).
const (
	RBR  = 0x00 // Receive Buffer Register
	THR  = 0x00 // Transmit Holding Register (same offset)
	IER  = 0x04 // Interrupt Enable Register
	FCR  = 0x08 // FIFO Control Register
	LCR  = 0x0c // Line Control Register
	MCR  = 0x10 // Modem Control Register
	FSR  = 0x18 // FIFO Status Register
	BAUD = 0x24 // Baud Rate Divisor Register
)

// Register bit positions.
const (
	LCR_WLS    = 0  // word length select [1:0]
	FCR_RFR    = 1  // RX FIFO reset
	FCR_TFR    = 2  // TX FIFO reset
	FSR_TXFULL = 23 // TX FIFO full
)

// WLS_8BIT selects an 8-bit word length in the LCR_WLS field.
const WLS_8BIT = 0b11

// UART represents a Nuvoton UART instance.
type UART struct {
	// Base register
	Base uint32
	// Baud is the value written to the BAUD divisor register, it encodes
	// both the divider mode and divisor for the configured input clock.
	Baud uint32
}

// Init configures the UART for 8N1 operation at the divisor set in Baud.
//
// Pin mux and the UART APB clock must be configured by the caller (e.g. in
// the board cpuinit) before Init.
func (hw *UART) Init() {
	if hw.Base == 0 {
		panic("invalid UART instance")
	}

	// 8-bit word length, no parity, 1 stop bit
	reg.SetN(hw.Base+LCR, LCR_WLS, 0b11, WLS_8BIT)

	// reset the TX/RX FIFOs
	reg.Set(hw.Base+FCR, FCR_RFR)
	reg.Set(hw.Base+FCR, FCR_TFR)

	// baud rate divisor
	reg.Write(hw.Base+BAUD, hw.Baud)
}

// Tx transmits a single byte, blocking until the TX FIFO has room.
func (hw *UART) Tx(c byte) {
	for reg.Get(hw.Base+FSR, FSR_TXFULL) {
	}
	reg.Write(hw.Base+THR, uint32(c))
}

// Write transmits buf.
func (hw *UART) Write(buf []byte) {
	for _, c := range buf {
		hw.Tx(c)
	}
}
