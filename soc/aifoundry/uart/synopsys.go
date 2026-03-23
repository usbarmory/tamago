// CORE-ET Silicom Platform Universal Asynchronous Receiver/Transmitter (UART) drivers
// https://github.com/usbarmory/tamago
//
// Copyright (c) The kotama Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package uart implements drivers for UART controllers found in CORE-ET
// Silicom Platform processors, adopting the following reference
// specifications:
//   - Shakti SoC Device Register Manual - 2021/06/24 - (Erbium)
//   - Synopsys DesignWare APB UART      - 2006/01/20 - (ET-SoC1)
//
// This package is only meant to be used with `GOOS=tamago GOARCH=riscv64` as
// supported by the TamaGo framework for bare metal Go on RISC-V SoCs, see
// https://github.com/usbarmory/tamago.
package uart

import (
	"github.com/usbarmory/tamago/internal/reg"
)

// APB UART registers
const (
	RBR_THR_DLL = 0x00
	DLH_IER     = 0x04

	IIR_FCR    = 0x08
	FCR_TFIFOR = 1
	FCR_RFIFOR = 1
	FCR_FIFOE  = 0

	LCR      = 0x0c
	LCR_DLAB = 7
	LCR_PEN  = 3
	LCR_STOP = 2
	LCR_DLS  = 0

	USR      = 0x7c
	USR_RFNE = 8
	USR_TFNF = 1

	SRR    = 0x88
	SRR_UR = 0
)

// APB represents a Synopsys APB UART serial port instance.
type Synopsys struct {
	// Controller index
	Index int
	// Base register
	Base uint32

	// control registers
	fifo   uint32
	status uint32
}

// SetBaudRate configures the divisor for 115200 baud
// clockHz is the frequency of the UART peripheral clock (e.g., 24000000)
func (hw *Synopsys) SetBaudRate(clockHz uint32, baud int) {
	reg.Set(hw.Base+LCR, LCR_DLAB)
	defer reg.Clear(hw.Base+LCR, LCR_DLAB)

	divisor := uint32(clockHz / uint32(16*baud))
	reg.Write(hw.Base+RBR_THR_DLL, divisor&0xff)
	reg.Write(hw.Base+DLH_IER, (divisor>>8)&0xff)
}

// Init initializes an APB serial port instance.
func (hw *Synopsys) Init() {
	if hw.Base == 0 {
		panic("invalid UART controller instance")
	}

	hw.fifo = hw.Base + RBR_THR_DLL
	hw.status = hw.Base + USR

	// software reset
	reg.Set(hw.Base+SRR, SRR_UR)

	// FIFO Enable, Reset RX/TX
	reg.Set(hw.Base+IIR_FCR, FCR_FIFOE)
	reg.Set(hw.Base+IIR_FCR, FCR_RFIFOR)
	reg.Set(hw.Base+IIR_FCR, FCR_TFIFOR)

	// 8-bit data length, 1 stop bit, parity disabled
	reg.SetN(hw.Base+LCR, LCR_DLS, 0b11, 0b11)
	reg.Clear(hw.Base+LCR, LCR_STOP)
	reg.Clear(hw.Base+LCR, LCR_PEN)

	hw.SetBaudRate(25_000_000, 115200)
}

// Tx transmits a single character to the serial port.
func (hw *Synopsys) Tx(c byte) {
	for reg.Get(hw.status, USR_TFNF) {
		// wait for TX FIFO to have room for a character
	}

	reg.Write(hw.fifo, uint32(c))
}

// Rx receives a single character from the serial port.
func (hw *Synopsys) Rx() (c byte, valid bool) {
	if reg.Get(hw.status, USR_RFNE) {
		return
	}

	return byte(reg.Read(hw.fifo)), true
}

// Write data from buffer to serial port.
func (hw *Synopsys) Write(buf []byte) (n int, _ error) {
	for n = range buf {
		hw.Tx(buf[n])
	}

	return
}

// Read available data to buffer from serial port.
func (hw *Synopsys) Read(buf []byte) (n int, _ error) {
	var valid bool

	for n = range buf {
		buf[n], valid = hw.Rx()

		if !valid {
			break
		}
	}

	return
}
