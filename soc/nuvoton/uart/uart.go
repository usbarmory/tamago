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

// LCR: 8-bit word length, no parity, 1 stop bit.
const lcr8N1 = 0x3

// FCR: reset and enable the TX/RX FIFOs.
const fcrFIFOEnable = 0x6

// fsrTXFull is set when the TX FIFO is full.
const fsrTXFull = 1 << 23

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

	reg.Write(hw.Base+LCR, lcr8N1)
	reg.Write(hw.Base+FCR, fcrFIFOEnable)
	reg.Write(hw.Base+BAUD, hw.Baud)
}

// Tx transmits a single byte, blocking until the TX FIFO has room.
func (hw *UART) Tx(c byte) {
	for reg.Read(hw.Base+FSR)&fsrTXFull != 0 {
	}
	reg.Write(hw.Base+THR, uint32(c))
}

// Write transmits buf.
func (hw *UART) Write(buf []byte) {
	for _, c := range buf {
		hw.Tx(c)
	}
}
