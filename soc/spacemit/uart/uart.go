// SpacemiT Universal Asynchronous Receiver/Transmitter (UART) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package uart implements a driver for SpacemiT UART controllers adopting the
// following reference specifications:
//   - Key Stone K1 Datasheet - V7.1 - 2026/08/31
//
// The controllers are 16550A/16750 compatible (p62, Section 2.7.7 UART
// Interface, K1 Datasheet) with 64-byte transmit and receive FIFOs. The
// register file is mapped with a 32-bit stride, therefore the standard 16550
// register indices are shifted left by two.
//
// Compatibility is however not complete, the controllers inherit Marvell
// PXA/MMP behaviour in two respects which a 16550 driver does not expect:
//
//   - the Interrupt Enable Register holds the unit enable, see [IER_POLLED]
//   - the divisor latch must be programmed high byte first, see
//     [UART.SetBaudRate]
//
// This package is only meant to be used with `GOOS=tamago GOARCH=riscv64` as
// supported by the TamaGo framework for bare metal Go on RISC-V SoCs, see
// https://github.com/usbarmory/tamago.
package uart

import "runtime"

// Defined in reg.s. One can not use internal/reg implementation,
// since these use load-reserved atomic instructions, which fail
// for MMIO address space on this SoC.
func Read(addr uint32) uint32
func Read64(addr uint64) uint64
func Write(addr uint32, val uint32)

// UART registers
const (
	UART_DEFAULT_BAUDRATE = 115200

	// Receive Buffer Register (read), Transmit Holding Register (write),
	// Divisor Latch Low (when LCR_DLAB is set).
	UARTx_RBR = 0x0000
	UARTx_THR = 0x0000
	UARTx_DLL = 0x0000

	// Interrupt Enable Register, Divisor Latch High (when LCR_DLAB is set).
	//
	// The register deviates from the 16550, inheriting the Marvell PXA/MMP
	// layout, in that it also holds the unit enable (UUE) along with the
	// NRZ coding and DMA request enables. Clearing it to disable interrupt
	// sources, as a 16550 driver would, disables the controller itself and
	// silences the port, see [IER_POLLED].
	//
	// The vendor boot loader programs the same value, through the
	// CONFIG_SYS_NS16550_IER=0x40 setting of its ns16550 driver.
	UARTx_IER = 0x0004
	UARTx_DLH = 0x0004
	IER_DMAE  = 7
	IER_UUE   = 6
	IER_NRZE  = 5
	IER_RTOIE = 4

	// IER value enabling the controller with all interrupt sources
	// disabled, as required for polled operation
	IER_POLLED = 1 << IER_UUE

	// Interrupt Identification Register (read), FIFO Control Register
	// (write).
	UARTx_IIR  = 0x0008
	UARTx_FCR  = 0x0008
	FCR_FIFOE  = 0
	FCR_RFIFOR = 1
	FCR_TFIFOR = 2

	// FCR value enabling the FIFOs and resetting both of them
	FCR_ENABLE = 1<<FCR_FIFOE | 1<<FCR_RFIFOR | 1<<FCR_TFIFOR

	// Line Control Register
	UARTx_LCR = 0x000c
	LCR_DLAB  = 7
	LCR_BC    = 6
	LCR_EPS   = 4
	LCR_PEN   = 3
	LCR_STOP  = 2
	LCR_DLS   = 0
	DLS_8     = 0b11

	// LCR value selecting 8 data bits, no parity, 1 stop bit
	LCR_8N1 = DLS_8 << LCR_DLS

	// Modem Control Register
	UARTx_MCR = 0x0010

	// Line Status Register
	UARTx_LSR = 0x0014
	LSR_TEMT  = 6
	LSR_THRE  = 5
	LSR_DR    = 0

	// Modem Status Register
	UARTx_MSR = 0x0018

	// Scratch Pad Register, a read/write location with no side effects on
	// controller operation.
	UARTx_SPR = 0x001c
)

// UART represents a serial port instance.
type UART struct {
	Index int
	Base  uint32
	// Clock returns the baud rate generator input clock frequency in Hz;
	// when set the divisor latch is programmed during Init (otherwise the
	// configuration left by an earlier boot stage is kept).
	Clock func() uint32
	// Baudrate is the desired baud rate; defaults to UART_DEFAULT_BAUDRATE
	// when zero. Only used when Clock is set.
	Baudrate uint32
}

// SetBaudRate configures the baud rate generator divisor latch for the
// argument input clock and baud rate, the divisor is computed as
// `round(clock / (16 * baudrate))`.
func (hw *UART) SetBaudRate(clock uint32, baudrate uint32) {
	if clock == 0 || baudrate == 0 {
		return
	}

	// divisor rounded to the nearest integer
	div := clock + 8*baudrate
	div /= 16 * baudrate

	if div == 0 {
		div = 1
	}

	lcr := Read(hw.Base + UARTx_LCR)
	Write(hw.Base+UARTx_LCR, lcr|1<<LCR_DLAB)

	// The controller requires the divisor latch to be programmed high byte
	// first, with a read back separating the two writes, rather than in the
	// 16550 low to high order ('ns16550_setbrg', drivers/serial/ns16550.c
	// under CONFIG_TARGET_SPACEMIT_K1X in the vendor U-Boot).
	Write(hw.Base+UARTx_DLH, (div>>8)&0xff)
	Read(hw.Base + UARTx_DLH)
	Write(hw.Base+UARTx_DLL, div&0xff)

	Write(hw.Base+UARTx_LCR, lcr & ^uint32(1<<LCR_DLAB))
}

// Divisor returns the baud rate generator divisor latch value, which before
// UART.Init is the one left by an earlier boot stage.
//
// It allows the input clock to be derived from a known good configuration, as
// `clock = divisor * 16 * baudrate`, rather than assumed.
func (hw *UART) Divisor() uint32 {
	lcr := Read(hw.Base + UARTx_LCR)

	Write(hw.Base+UARTx_LCR, lcr|1<<LCR_DLAB)

	dll := Read(hw.Base+UARTx_DLL) & 0xff
	dlh := Read(hw.Base+UARTx_DLH) & 0xff

	Write(hw.Base+UARTx_LCR, lcr)

	return dlh<<8 | dll
}

// Init initializes and enables the UART for 8N1 polled operation.
func (hw *UART) Init() {
	if hw.Base == 0 {
		panic("invalid UART controller instance")
	}

	// an earlier boot stage might have left the divisor latch selected
	lcr := Read(hw.Base + UARTx_LCR)
	Write(hw.Base+UARTx_LCR, lcr & ^uint32(1<<LCR_DLAB))

	// The driver is polling based, disable all interrupt sources while
	// keeping the controller enabled, see IER_POLLED.
	Write(hw.Base+UARTx_IER, IER_POLLED)

	if hw.Clock != nil {
		baudrate := hw.Baudrate

		if baudrate == 0 {
			baudrate = UART_DEFAULT_BAUDRATE
		}

		hw.SetBaudRate(hw.Clock(), baudrate)
	}

	// 8 data bits, no parity, 1 stop bit
	Write(hw.Base+UARTx_LCR, LCR_8N1)

	// enable and reset both FIFOs
	Write(hw.Base+UARTx_FCR, FCR_ENABLE)
}

// TryTx attempts to transmit a single character, polling the TX FIFO at most
// attempts times. It returns false without writing when the FIFO remains full.
func (hw *UART) TryTx(c byte, attempts int) bool {
	for attempts > 0 {
		if Read(hw.Base+UARTx_LSR)&(1<<LSR_THRE) != 0 {
			Write(hw.Base+UARTx_THR, uint32(c))
			return true
		}

		attempts--
	}

	return false
}

// TrySyncTx transmits a single character and waits for the transmitter to go
// completely idle, polling at most attempts times for each step. It reports
// whether the character was both accepted and fully shifted out.
//
// LSR_THRE only reports the transmit holding register empty, not the FIFO
// drained, so [UART.TryTx] returns as soon as a character is queued. A caller
// emitting at full speed can therefore have characters accepted that are never
// observed on the wire, which makes lost output indistinguishable from a
// program that stopped running. Diagnostic paths, where that distinction is
// the whole point, should prefer this method.
func (hw *UART) TrySyncTx(c byte, attempts int) bool {
	if !hw.TryTx(c, attempts) {
		return false
	}

	for i := 0; i < attempts; i++ {
		if Read(hw.Base+UARTx_LSR)&(1<<LSR_TEMT) != 0 {
			return true
		}
	}

	return false
}

// Tx transmits a single character to the serial port.
func (hw *UART) Tx(c byte) {
	for !hw.TryTx(c, 1) {
		// wait for TX FIFO to have room for a character
	}
}

// Rx receives a single character from the serial port.
func (hw *UART) Rx() (c byte, valid bool) {
	if Read(hw.Base+UARTx_LSR)&(1<<LSR_DR) == 0 {
		return
	}

	return byte(Read(hw.Base+UARTx_RBR) & 0xff), true
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
