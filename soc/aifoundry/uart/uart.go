// Shakti Universal Asynchronous Receiver/Transmitter (UART) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The kotama Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package uart implements a driver for UART controllers adopting the following
// reference specifications:
//   - Shakti SoC Device Register Manual - 2021/06/24
//
// This package is only meant to be used with `GOOS=tamago GOARCH=riscv64` as
// supported by the TamaGo framework for bare metal Go on RISC-V SoCs, see
// https://github.com/usbarmory/tamago.
package uart

import (
	"github.com/usbarmory/tamago/internal/reg"
)

// UART registers
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

// UART represents a serial port instance.
type UART struct {
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
func (hw *UART) Init() {
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
func (hw *UART) Tx(c byte) {
	for reg.Get(hw.status, STATUS_TX_FULL) {
		// wait for TX FIFO to have room for a character
	}

	reg.Write(hw.txdata, uint32(c))
}

// Rx receives a single character from the serial port.
func (hw *UART) Rx() (c byte, valid bool) {
	if !reg.Get(hw.status, STATUS_RX_READY) {
		return
	}

	return byte(reg.Read(hw.rxdata)), true
}

// Write data from buffer to serial port.
func (hw *UART) Write(buf []byte) (n int, _ error) {
	for n = range buf {
		hw.Tx(buf[n])
	}

	return
}

// Read available data to buffer from serial port.
func (hw *UART) Read(buf []byte) (n int, _ error) {
	var valid bool

	for n = range buf {
		buf[n], valid = hw.Rx()

		if !valid {
			break
		}
	}

	return
}
