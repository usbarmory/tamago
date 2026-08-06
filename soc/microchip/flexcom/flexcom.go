// Microchip Flexible Serial Communication Controller (FLEXCOM)
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package flexcom implements a driver for Flexible Serial Communication
// Controllers (FLEXCOM) in USART and TWI initiator modes.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm64` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package flexcom

import (
	"math"
	"runtime"
	"sync"
	"time"

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

	FLEX_MR         = 0x00
	MR_OPMODE       = 0
	MR_OPMODE_MASK  = 0x3
	MR_OPMODE_USART = 1
	MR_OPMODE_TWI   = 3

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

// FLEXCOM represents a Flexible Serial Communication controller instance.
type FLEXCOM struct {
	sync.Mutex

	// Controller index
	Index int
	// Base register
	Base uint32
	// Baud rate
	Baudrate uint32
	// Interrupt ID
	IRQ int

	// TWIClockLowDivider configures the TWI low-period divider.
	TWIClockLowDivider uint8
	// TWIClockHighDivider configures the TWI high-period divider.
	TWIClockHighDivider uint8
	// TWIClockDivider applies a 2^n scale to both TWI clock periods.
	TWIClockDivider uint8
	// TWIGenericClock selects GCLK instead of the peripheral clock.
	TWIGenericClock bool
	// TWITimeout bounds each wait for TWI controller progress. Zero selects 100 ms.
	TWITimeout time.Duration

	// flexcom control register
	mr uint32

	// usart control registers
	us_cr   uint32
	us_mr   uint32
	us_ier  uint32
	us_idr  uint32
	us_csr  uint32
	us_rhr  uint32
	us_thr  uint32
	us_brgr uint32
	us_fmr  uint32

	twiClock uint32

	rx chan bool
}

// Init initializes and enables a FLEXCOM controller instance in USART mode.
func (hw *FLEXCOM) Init() {
	hw.Lock()
	defer hw.Unlock()

	if hw.Base == 0 {
		panic("invalid FLEXCOM controller instance")
	}

	if hw.Baudrate == 0 {
		hw.Baudrate = UART_BAUDRATE_DEFAULT
	}

	hw.mr = hw.Base + FLEX_MR
	hw.us_cr = hw.Base + FLEX_USART_OFFSET + FLEX_US_CR
	hw.us_mr = hw.Base + FLEX_USART_OFFSET + FLEX_US_MR
	hw.us_ier = hw.Base + FLEX_USART_OFFSET + FLEX_US_IER
	hw.us_idr = hw.Base + FLEX_USART_OFFSET + FLEX_US_IDR
	hw.us_csr = hw.Base + FLEX_USART_OFFSET + FLEX_US_CSR
	hw.us_rhr = hw.Base + FLEX_USART_OFFSET + FLEX_US_RHR
	hw.us_thr = hw.Base + FLEX_USART_OFFSET + FLEX_US_THR
	hw.us_brgr = hw.Base + FLEX_USART_OFFSET + FLEX_US_BRGR
	hw.us_fmr = hw.Base + FLEX_USART_OFFSET + FLEX_US_FMR

	hw.setup()
}

func (hw *FLEXCOM) setup() {
	// set USART operating mode
	reg.SetN(hw.mr, MR_OPMODE, MR_OPMODE_MASK, MR_OPMODE_USART)

	// set baud rate
	// p583, 3.47.11.2.5 USART Functional Description, Microchip DS00005048E
	//              SelectedClock
	// baudrate = -----------------
	//            8 * (2-OVER) * CD
	//
	cd := math.Round(float64(PERIPHERAL_CLK) / (16 * float64(hw.Baudrate)))
	reg.Write(hw.us_brgr, uint32(cd))

	// reset the receiver and transmitter
	reg.Write(hw.us_cr, 1<<US_CR_RSTRX)
	reg.Write(hw.us_cr, 1<<US_CR_RSTTX)
	reg.Write(hw.us_cr, 1<<US_CR_RXDIS)
	reg.Write(hw.us_cr, 1<<US_CR_TXDIS)

	// set 8N1 mode
	reg.SetN(hw.us_mr, US_MR_PAR, 0b111, 4)
	reg.SetN(hw.us_mr, US_MR_CHRL, 0b11, 3)
	reg.SetN(hw.us_mr, US_MR_NBSTOP, 0b11, 0)

	// set asynchronous mode (UART)
	reg.ClearN(hw.us_mr, US_MR_SYNC, 1)

	// single-data FIFOs
	reg.Write(hw.us_fmr, 0)
	reg.Write(hw.us_cr, 1<<US_CR_FIFOEN)
	reg.Write(hw.us_cr, 1<<US_CR_RXFCLR)
	reg.Write(hw.us_cr, 1<<US_CR_TXFCLR)

	// enable Tx and RX
	reg.Write(hw.us_cr, 1<<US_CR_RXEN)
	reg.Write(hw.us_cr, 1<<US_CR_TXEN)
}

// EnableInterrupt enables interrupt generation for receive FIFOs. Once enabled
// [FLEXCOM.Read] and [FLEXCOM.Rx] block, as required, on the argument channel
// rather than polling for valid data. A nil channel disables receive
// interrupts and restores polling.
func (hw *FLEXCOM) EnableInterrupt(rx chan bool) {
	if rx == nil {
		reg.Write(hw.us_idr, 1<<US_IER_RXRDY_IE)
	} else {
		reg.Write(hw.us_ier, 1<<US_IER_RXRDY_IE)
	}

	hw.rx = rx
}

// Tx transmits a single character to the serial port.
func (hw *FLEXCOM) Tx(c byte) {
	for reg.GetN(hw.us_csr, US_CSR_TXRDY, 1) == 0 {
		// wait for TX FIFO to have room for a character
	}

	reg.Write8(hw.us_thr, c)
}

// Rx receives a single character from the serial port, waiting for data to
// become available if the argument is true.
func (hw *FLEXCOM) Rx(block bool) (c byte, valid bool) {
	for {
		if reg.GetN(hw.us_csr, US_CSR_RXRDY, 1) == 1 {
			return reg.Read8(hw.us_rhr), true
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
func (hw *FLEXCOM) Write(buf []byte) (n int, _ error) {
	for _, c := range buf {
		hw.Tx(c)
	}

	return len(buf), nil
}

// Read available data to buffer from serial port.
func (hw *FLEXCOM) Read(buf []byte) (n int, _ error) {
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
