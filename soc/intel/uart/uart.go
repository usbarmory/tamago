// 16550 Universal Asynchronous Receiver/Transmitter (UART) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package uart implements a driver for Intel Serial I/O UART controllers adopting the
// following reference specifications:
//   - PC16550D - Universal Asynchronous Receiver/Transmitter with FIFOs - June 1995
//
// This package is only meant to be used with `GOOS=tamago GOARCH=amd64` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package uart

import (
	"runtime"

	"github.com/usbarmory/tamago/internal/reg"
)

// UART registers
const (
	DEFAULT_BAUDRATE = 115200

	RBR = 0x00
	THR = 0x00

	IER       = 0x01
	IER_ERBFI = 0

	FCR = 0x02
	LCR = 0x03

	MCR       = 0x04
	MCR_INTEN = 3
	MCR_RTS   = 1
	MCR_DTR   = 0

	LSR      = 0x05
	LSR_DR   = 0
	LSR_THRE = 5

	MSR = 0x06
)

// UART represents a serial port instance.
type UART struct {
	// Controller index
	Index int
	// Base register
	Base uint16
	// Interrupt ID
	IRQ int

	// Data Terminal Ready
	DTR bool
	// Request To Send
	RTS bool

	rx chan bool
}

// Init initializes and enables the UART.
func (hw *UART) Init() {
	if hw.Base == 0 {
		panic("invalid UART controller instance")
	}

	mcr := uint8(1 << MCR_INTEN)

	if hw.RTS {
		mcr |= 1 << MCR_RTS
	}

	if hw.DTR {
		mcr |= 1 << MCR_DTR
	}

	reg.Out8(hw.Base+MCR, mcr)
}

// EnableInterrupt enables interrupt generation for receive FIFOs. Once enabled
// [UART.Read] and [UART.Rx] block, as required, on the argument channel rather
// than polling for valid data.
func (hw *UART) EnableInterrupt(rx chan bool) {
	reg.Out8(hw.Base+IER, 1 << IER_ERBFI)
	hw.rx = rx
}

// Tx transmits a single character to the serial port.
func (hw *UART) Tx(c byte) {
	for reg.In8(hw.Base+LSR)&(1<<LSR_THRE) == 0 {
		// wait for TX FIFO to have room for a character
	}

	reg.Out8(hw.Base+THR, uint8(c))
}

// Rx receives a single character from the serial port.
func (hw *UART) Rx(block bool) (c byte, valid bool) {
	if hw.rx != nil {
		if block {
			<-hw.rx
		} else {
			select {
			case <-hw.rx:
			default:
				return
			}
		}
	}

	if reg.In8(hw.Base+LSR)&(1<<LSR_DR) == 1 {
		return byte(reg.In8(hw.Base + RBR)), true
	}

	if block && hw.rx == nil {
		for reg.In8(hw.Base+LSR)&(1<<LSR_DR) == 0 {
			runtime.Gosched()
		}

		return byte(reg.In8(hw.Base + RBR)), true
	}

	return
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
