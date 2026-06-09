// SiFive Universal Asynchronous Receiver/Transmitter (UART) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
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
	"runtime"

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
	CTRL_EN      = 0

	// baud rate divisor: fbaud = Clock / (div + 1)
	UARTx_DIV = 0x0018

	// Framing setup register (Nuclei UX600 variant only); the standard
	// SiFive FU540 UART does not implement it. UART_SETUP_8N1 selects 8N1.
	UARTx_SETUP    = 0x0020
	UART_SETUP_8N1 = 0x30
)

// UART represents a serial port instance.
type UART struct {
	// Controller index
	Index int
	// Base register
	Base uint32
	// Clock returns the UART input clock frequency in Hz; when set the baud
	// rate divisor is programmed during Init (otherwise the divisor left by
	// an earlier boot stage is kept).
	Clock func() uint32
	// Baudrate is the desired baud rate; defaults to UART_DEFAULT_BAUDRATE
	// when zero. Only used when Clock is set.
	Baudrate uint32
	// Setup, when non-zero, is written to the framing setup register
	// (offset 0x20) during Init. Required by the Nuclei UX600 UART variant
	// (UART_SETUP_8N1); leave zero for standard SiFive FU540 UARTs.
	Setup uint32

	// control registers
	txdata uint32
	rxdata uint32
}

// Init initializes and enables the UART for RS-232 mode,
// p3605, 55.13.1 Programming the UART in RS-232 mode, IMX6ULLRM.
func (hw *UART) Init() {
	if hw.Base == 0 {
		panic("invalid UART controller instance")
	}

	hw.txdata = hw.Base + UARTx_TXDATA
	hw.rxdata = hw.Base + UARTx_RXDATA

	if hw.Clock != nil {
		baudrate := hw.Baudrate

		if baudrate == 0 {
			baudrate = UART_DEFAULT_BAUDRATE
		}

		// div = ceil(f_in / baud) - 1 (SiFive FSBL convention, matching
		// the vendor OpenSBI and U-Boot drivers).
		clock := hw.Clock()
		reg.Write(hw.Base+UARTx_DIV, (clock+baudrate-1)/baudrate-1)
	}

	if hw.Setup != 0 {
		reg.Write(hw.Base+UARTx_SETUP, hw.Setup)
	}

	reg.Set(hw.Base+UARTx_TXCTRL, CTRL_EN)
	reg.Set(hw.Base+UARTx_RXCTRL, CTRL_EN)
}

// Tx transmits a single character to the serial port.
func (hw *UART) Tx(c byte) {
	for reg.GetN(hw.txdata, TXDATA_FULL, 1) == 1 {
		// wait for TX FIFO to have room for a character
	}

	reg.Write(hw.txdata, uint32(c))
}

// Rx receives a single character from the serial port.
func (hw *UART) Rx() (c byte, valid bool) {
	rxdata := reg.Read(hw.rxdata)

	if bits.GetN(&rxdata, RXDATA_EMPTY, 1) == 1 {
		return
	}

	return byte(bits.GetN(&rxdata, RXDATA_DATA, 0xff)), true
}

// Write data from buffer to serial port.
func (hw *UART) Write(buf []byte) (n int, _ error) {
	for _, c := range buf {
		hw.Tx(c)
	}

	return len(buf), nil
}

// Read available data to buffer from serial port.
func (hw *UART) Read(buf []byte) (n int, _ error) {
	var valid bool

	for n < len(buf) {
		buf[n], valid = hw.Rx()

		if !valid {
			if n == 0 {
				runtime.Gosched()
			}

			break
		}

		n++
	}

	return
}
