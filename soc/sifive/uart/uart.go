// SiFive UART driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package uart implements a driver for SiFive UART controllers adopting the
// following reference specifications:
//   - FU540C00RM - SiFive FU540-C000 Manual - v1p4 2021/03/25
//
// This package is only meant to be used with `GOOS=tamago GOARCH=riscv64` as
// supported by the TamaGo framework for bare metal Go on RISC-V SoCs, see
// https://github.com/usbarmory/tamago.
package uart

import (
	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
)

// UART registers
const (
	UART_DEFAULT_BAUDRATE = 115200

	// p94, Chapter 13 UART, FU540C00RM

	UARTx_TXDATA = 0x0000
	TXDATA_FULL  = 31
	TXDATA_DATA  = 0

	UARTx_RXDATA = 0x0004
	RXDATA_EMPTY = 31
	RXDATA_DATA  = 0

	UARTx_TXCTRL = 0x0008
	UARTx_RXCTRL = 0x000c
)

// UART represents a serial port instance.
type UART struct {
	// Controller index
	Index int
	// Base register
	Base uint32

	// control registers
	txdata uint32
	rxdata uint32
	txctrl uint32
	rxctrl uint32
}

// Init initializes and enables the UART for RS-232 mode,
// p3605, 55.13.1 Programming the UART in RS-232 mode, IMX6ULLRM.
func (hw *UART) Init() {
	if hw.Base == 0 {
		panic("invalid UART controller instance")
	}

	hw.txdata = hw.Base + UARTx_TXDATA
	hw.rxdata = hw.Base + UARTx_RXDATA
	hw.txctrl = hw.Base + UARTx_TXCTRL
	hw.rxctrl = hw.Base + UARTx_RXCTRL
}

func (hw *UART) txFull() bool {
	return reg.Get(hw.txdata, TXDATA_FULL, 1) == 1
}

// Tx transmits a single character to the serial port.
func (hw *UART) Tx(c byte) {
	for hw.txFull() {
		// wait for TX FIFO to have room for a character
	}
	reg.Write(hw.txdata, uint32(c))
}

// Rx receives a single character from the serial port.
func (hw *UART) Rx() (c byte, valid bool) {
	rxdata := reg.Read(hw.rxdata)

	if bits.Get(&rxdata, RXDATA_EMPTY, 1) == 1 {
		return
	}

	return byte(bits.Get(&rxdata, RXDATA_DATA, 0xff)), true
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
