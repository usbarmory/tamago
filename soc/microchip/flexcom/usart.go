// Microchip FLEXCOM USART support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package flexcom

import (
	"math"
	"runtime"

	"github.com/usbarmory/tamago/internal/reg"
)

// USART registers
const (
	UART_BAUDRATE_DEFAULT = 115200
	UART_BAUDRATE_HS      = 921600

	FLEX_USART_OFFSET = 0x200

	FLEX_US_CR    = 0x00
	US_CR_FIFODIS = 31
	US_CR_FIFOEN  = 30
	US_CR_RXFCLR  = 25
	US_CR_TXFCLR  = 24
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

	FLEX_US_IER     = 0x08
	US_IER_TXRDY_IE = 1
	US_IER_RXRDY_IE = 0

	FLEX_US_IDR = 0x0c

	FLEX_US_CSR  = 0x14
	US_CSR_TXRDY = 1
	US_CSR_RXRDY = 0

	FLEX_US_RHR = 0x18
	FLEX_US_THR = 0x1c

	FLEX_US_BRGR = 0x20
	US_BRGR_CD   = 0

	FLEX_US_FMR = 0xa0
)

// USART represents the FLEXCOM USART mode instance.
type USART struct {
	// Baud rate
	Baudrate uint32

	// usart control registers
	cr   uint32
	mr   uint32
	ier  uint32
	idr  uint32
	csr  uint32
	rhr  uint32
	thr  uint32
	brgr uint32
	fmr  uint32

	rx chan bool
}

func (hw *USART) init() {
	if hw.Baudrate == 0 {
		hw.Baudrate = UART_BAUDRATE_DEFAULT
	}

	// set baud rate
	// p583, 3.47.11.2.5 USART Functional Description, Microchip DS00005048E
	//              SelectedClock
	// baudrate = -----------------
	//            8 * (2-OVER) * CD
	//
	cd := math.Round(float64(PERIPHERAL_CLK) / (16 * float64(hw.Baudrate)))
	reg.Write(hw.brgr, uint32(cd))

	// reset the receiver and transmitter
	reg.Write(hw.cr, 1<<US_CR_RSTRX)
	reg.Write(hw.cr, 1<<US_CR_RSTTX)
	reg.Write(hw.cr, 1<<US_CR_RXDIS)
	reg.Write(hw.cr, 1<<US_CR_TXDIS)

	// set 8N1 mode
	reg.SetN(hw.mr, US_MR_PAR, 0b111, 4)
	reg.SetN(hw.mr, US_MR_CHRL, 0b11, 3)
	reg.SetN(hw.mr, US_MR_NBSTOP, 0b11, 0)

	// set asynchronous mode (UART)
	reg.ClearN(hw.mr, US_MR_SYNC, 1)

	// single-data FIFOs
	reg.Write(hw.fmr, 0)
	reg.Write(hw.cr, 1<<US_CR_FIFOEN)
	reg.Write(hw.cr, 1<<US_CR_RXFCLR)
	reg.Write(hw.cr, 1<<US_CR_TXFCLR)

	// enable Tx and RX
	reg.Write(hw.cr, 1<<US_CR_RXEN)
	reg.Write(hw.cr, 1<<US_CR_TXEN)
}

// EnableInterrupt enables interrupt generation for receive FIFOs. Once enabled
// [USART.Read] and [USART.Rx] block, as required, on the argument channel
// rather than polling for valid data. A nil channel disables receive
// interrupts and restores polling.
func (hw *USART) EnableInterrupt(rx chan bool) {
	if rx == nil {
		reg.Write(hw.idr, 1<<US_IER_RXRDY_IE)
	} else {
		reg.Write(hw.ier, 1<<US_IER_RXRDY_IE)
	}

	hw.rx = rx
}

// Tx transmits a single character to the serial port.
func (hw *USART) Tx(c byte) {
	for reg.GetN(hw.csr, US_CSR_TXRDY, 1) == 0 {
		// wait for TX FIFO to have room for a character
	}

	reg.Write8(hw.thr, c)
}

// Rx receives a single character from the serial port, waiting for data to
// become available if the argument is true.
func (hw *USART) Rx(block bool) (c byte, valid bool) {
	for {
		if reg.GetN(hw.csr, US_CSR_RXRDY, 1) == 1 {
			return reg.Read8(hw.rhr), true
		}

		if !block {
			return
		}

		if hw.rx != nil {
			<-hw.rx
		} else {
			runtime.Gosched()
		}
	}
}

// Write data from buffer to serial port.
func (hw *USART) Write(buf []byte) (n int, _ error) {
	for _, c := range buf {
		hw.Tx(c)
	}

	return len(buf), nil
}

// Read available data to buffer from serial port.
func (hw *USART) Read(buf []byte) (n int, _ error) {
	block := true

	for n < len(buf) {
		c, valid := hw.Rx(block)

		if !valid {
			break
		}

		buf[n] = c
		n++

		if n == 1 {
			block = false
		}
	}

	return
}
