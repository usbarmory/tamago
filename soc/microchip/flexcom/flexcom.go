// Microchip Flexible Serial Communication Controller (FLEXCOM)
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package flexcom implements a driver for Flexible Serial Communication
// Controllers (FLEXCOM), currently only USART mode is supported.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm64` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package flexcom

import (
	"math"
	"runtime"

	"github.com/usbarmory/tamago/internal/reg"
)

// FLEXCOM registers
const (
	UART_BAUDRATE_DEFAULT = 115200
	UART_BAUDRATE_HS      = 921600

	// Peripheral clock is 250 MHz
	// p501, 3.47.3 CLOCKING AND RESET, Microchip DS00005048E
	PERIPHERAL_CLK = 250e6

	FLEX_USART_OFFSET = 0x200

	FLEX_MR   = 0x00
	MR_OPMODE = 0

	FLEX_US_CR    = 0x00
	US_CR_FIFODIS = 31
	US_CR_FIFOEN  = 30
	US_CR_TXDIS   = 7
	US_CR_TXEN    = 6
	US_CR_RXDIS   = 5
	US_CR_RXEN    = 4
	US_CR_RSTTX   = 3
	US_CR_RSTRX   = 2

	FLEX_US_MR   = 0x04
	US_MR_CHMODE = 14
	US_MR_NBSTOP = 12
	US_MR_PAR    = 9
	US_MR_SYNC   = 8
	US_MR_CHRL   = 6

	FLEX_US_CSR  = 0x14
	US_CSR_TXRDY = 1
	US_CSR_RXRDY = 0

	FLEX_US_RHR = 0x18
	FLEX_US_THR = 0x1c

	FLEX_US_BRGR = 0x20
	US_BRGR_CD   = 0
)

// FLEXCOM represents a Flexible Serial Communication controller instance.
type FLEXCOM struct {
	// Controller index
	Index int
	// Base register
	Base uint32
	// Baud rate
	Baudrate uint32

	// flexcom control register
	mr uint32

	// usart control registers
	us_cr   uint32
	us_mr   uint32
	us_csr  uint32
	us_rhr  uint32
	us_thr  uint32
	us_brgr uint32
}

// Init initializes and enables an FLEXCOM controller instance in USART mode.
func (hw *FLEXCOM) Init() {
	if hw.Base == 0 {
		panic("invalid FLEXCOM controller instance")
	}

	if hw.Baudrate == 0 {
		hw.Baudrate = UART_BAUDRATE_DEFAULT
	}

	hw.mr = hw.Base + FLEX_MR
	hw.us_cr = hw.Base + FLEX_USART_OFFSET + FLEX_US_CR
	hw.us_mr = hw.Base + FLEX_USART_OFFSET + FLEX_US_MR
	hw.us_csr = hw.Base + FLEX_USART_OFFSET + FLEX_US_CSR
	hw.us_rhr = hw.Base + FLEX_USART_OFFSET + FLEX_US_RHR
	hw.us_thr = hw.Base + FLEX_USART_OFFSET + FLEX_US_THR
	hw.us_brgr = hw.Base + FLEX_USART_OFFSET + FLEX_US_BRGR

	hw.setup()
}

func (hw *FLEXCOM) setup() {
	// set USART operating mode
	reg.SetN(hw.mr, MR_OPMODE, 0b11, 1)

	// set baud rate
	// p583, 3.47.11.2.5 USART Functional Description, Microchip DS00005048E
	//              SelectedClock
	// baudrate = -----------------
	//            8 * (2-OVER) * CD
	//
	cd := math.Round(float64(PERIPHERAL_CLK) / (16 * float64(hw.Baudrate)))
	reg.Write(hw.us_brgr, uint32(cd))

	// reset the receiver and transmitter
	reg.SetN(hw.us_cr, US_CR_RSTRX, 1, 1)
	reg.SetN(hw.us_cr, US_CR_RSTTX, 1, 1)
	reg.SetN(hw.us_cr, US_CR_RXDIS, 1, 1)
	reg.SetN(hw.us_cr, US_CR_TXDIS, 1, 1)

	// set 8N1 mode
	reg.SetN(hw.us_mr, US_MR_PAR, 0b111, 4)
	reg.SetN(hw.us_mr, US_MR_CHRL, 0b11, 3)
	reg.SetN(hw.us_mr, US_MR_NBSTOP, 0b11, 0)

	// set asynchronous mode (UART)
	reg.ClearN(hw.us_mr, US_MR_SYNC, 1)

	// enable Tx and RX
	reg.SetN(hw.us_cr, US_CR_RXEN, 1, 1)
	reg.SetN(hw.us_cr, US_CR_TXEN, 1, 1)
}

// Tx transmits a single character to the serial port.
func (hw *FLEXCOM) Tx(c byte) {
	for reg.GetN(hw.us_csr, US_CSR_TXRDY, 1) == 0 {
		// wait for TX FIFO to have room for a character
	}

	reg.Write(hw.us_thr, uint32(c))
}

// Rx receives a single character from the serial port.
func (hw *FLEXCOM) Rx() (c byte, valid bool) {
	if reg.GetN(hw.us_csr, US_CSR_RXRDY, 1) == 1 {
		return byte(reg.Read(hw.us_rhr)), true
	}

	return
}

// Write data from buffer to serial port.
func (hw *FLEXCOM) Write(buf []byte) (n int, _ error) {
	for n = range buf {
		hw.Tx(buf[n])
	}

	return
}

// Read available data to buffer from serial port.
func (hw *FLEXCOM) Read(buf []byte) (n int, _ error) {
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
