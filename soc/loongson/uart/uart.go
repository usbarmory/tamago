// NS16550-compatible UART driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package uart implements a driver for NS16550-compatible UART controllers,
// such as the legacy console found on Loongson LS7A based platforms and the
// QEMU LoongArch `virt` machine.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=loong64` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package uart

import (
	"runtime"

	"github.com/usbarmory/tamago/internal/reg"
)

// NS16550 register offsets (register shift 0, byte access).
const (
	RBR = 0x00 // Receiver Buffer Register (read, DLAB=0)
	THR = 0x00 // Transmitter Holding Register (write, DLAB=0)
	IER = 0x01 // Interrupt Enable Register
	FCR = 0x02 // FIFO Control Register
	LCR = 0x03 // Line Control Register

	LSR      = 0x05 // Line Status Register
	LSR_THRE = 5    // Transmitter Holding Register Empty
	LSR_DR   = 0    // Data Ready
)

// UART represents a serial port instance.
type UART struct {
	// Base register address
	Base uint32
}

// Init initializes the UART for 8N1 operation with FIFOs enabled. The baud
// rate divisor is left untouched as it is irrelevant under emulation.
func (hw *UART) Init() {
	reg.Write8(hw.Base+IER, 0x00) // disable interrupts
	reg.Write8(hw.Base+FCR, 0xc7) // enable and clear RX/TX FIFOs, 14-byte trigger
	reg.Write8(hw.Base+LCR, 0x03) // 8 bits, no parity, 1 stop bit
}

// Tx transmits a single character to the serial port.
func (hw *UART) Tx(c byte) {
	for reg.Read8(hw.Base+LSR)&(1<<LSR_THRE) == 0 {
		// wait for the transmitter holding register to drain
	}

	reg.Write8(hw.Base+THR, c)
}

// Rx receives a single character from the serial port, the second return value
// indicates whether a character was available.
func (hw *UART) Rx() (c byte, valid bool) {
	if reg.Read8(hw.Base+LSR)&(1<<LSR_DR) == 0 {
		return
	}

	return reg.Read8(hw.Base + RBR), true
}

// Write transmits the argument buffer to the serial port.
func (hw *UART) Write(buf []byte) (n int, _ error) {
	for _, c := range buf {
		hw.Tx(c)
	}

	return len(buf), nil
}

// Read receives the available bytes from the serial port into the argument
// buffer.
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
