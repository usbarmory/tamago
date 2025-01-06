// 16550A UART driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package microvm

import (
	"github.com/usbarmory/tamago/internal/reg"
)

// UART registers
const (
	DEFAULT_BAUDRATE = 115200

	RBR = 0x00
	THR = 0x00
	IER = 0x01
	FCR = 0x02
	MCR = 0x04

	LSR      = 0x05
	LSR_DR   = 0
	LSR_THRE = 5
)

// UART represents a serial port instance.
type UART struct {
	// Controller index
	Index int
	// Base register
	Base uint16
}

// Init initializes and enables the UART.
func (hw *UART) Init() {
	if hw.Base == 0 {
		panic("invalid UART controller instance")
	}
}

// Tx transmits a single character to the serial port.
func (hw *UART) Tx(c byte) {
	for reg.In8(hw.Base+LSR)&(1<<LSR_THRE) == 0 {
		// wait for TX FIFO to have room for a character
	}

	reg.Out8(hw.Base+THR, uint8(c))
}

// Rx receives a single character from the serial port.
func (hw *UART) Rx() (c byte, valid bool) {
	if reg.In8(hw.Base+LSR)&(1<<LSR_DR) == 0 {
		return
	}

	return byte(reg.In8(hw.Base + RBR)), true
}

// Write data from buffer to serial port.
func (hw *UART) Write(buf []byte) (n int, _ error) {
	for n = 0; n < len(buf); n++ {
		hw.Tx(buf[n])
	}

	return
}

// Read available data to buffer from serial port.
func (hw *UART) Read(buf []byte) (n int, _ error) {
	var valid bool

	for n = 0; n < len(buf); n++ {
		buf[n], valid = hw.Rx()

		if !valid {
			break
		}
	}

	return
}
