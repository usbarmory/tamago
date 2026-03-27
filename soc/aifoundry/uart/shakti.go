// CORE-ET Silicom Platform Universal Asynchronous Receiver/Transmitter (UART) drivers
// https://github.com/usbarmory/tamago
//
// Copyright (c) The kotama Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package uart

import (
	"runtime"

	"github.com/usbarmory/tamago/internal/reg"
)

// Shakti UART registers
const (
	CONFIG_UART_ENABLE = 6

	UART_BAUD = 0x00
	UART_TX   = 0x08
	UART_RX   = 0x10

	UART_STATUS     = 0x18
	STATUS_RX_READY = 2
	STATUS_TX_FULL  = 1
	STATUS_TX_EMPTY = 0
)

// Shakti represents a Shakti serial port instance.
type Shakti struct {
	// Controller index
	Index int
	// Base register
	Base uint32
	// System config register
	System uint32

	// control registers
	txdata uint32
	rxdata uint32
	status uint32
}

// Init initializes a serial port instance.
func (hw *Shakti) Init() {
	if hw.Base == 0 {
		panic("invalid UART controller instance")
	}

	hw.txdata = hw.Base + UART_TX
	hw.rxdata = hw.Base + UART_RX
	hw.status = hw.Base + UART_STATUS

	if hw.System != 0 {
		reg.Set(hw.System, CONFIG_UART_ENABLE)
	}
}

// Tx transmits a single character to the serial port.
func (hw *Shakti) Tx(c byte) {
	for reg.Get(hw.status, STATUS_TX_FULL) {
		// wait for TX FIFO to have room for a character
	}

	reg.Write(hw.txdata, uint32(c))
}

// Rx receives a single character from the serial port.
func (hw *Shakti) Rx() (c byte, valid bool) {
	if !reg.Get(hw.status, STATUS_RX_READY) {
		return
	}

	return byte(reg.Read(hw.rxdata)), true
}

// Write data from buffer to serial port.
func (hw *Shakti) Write(buf []byte) (n int, _ error) {
	for n = range buf {
		hw.Tx(buf[n])
	}

	return
}

// Read available data to buffer from serial port.
func (hw *Shakti) Read(buf []byte) (n int, _ error) {
	var valid bool

	for n = range buf {
		buf[n], valid = hw.Rx()

		if !valid {
			if n == 0 {
				runtime.Gosched()
			}

			break
		}
	}

	return
}
