// NXP UART driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package uart implements a driver for NXP UART controllers adopting the
// following reference specifications:
//   - IMX6ULLRM - i.MX 6ULL Applications Processor Reference Manual - Rev 1 2017/11
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/usbarmory/tamago.
package uart

import (
	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
)

// UART registers
const (
	UART_DEFAULT_BAUDRATE = 115200
	ESC                   = 0x1b

	// p3608, 55.15 UART Memory Map/Register Definition, IMX6ULLRM

	UARTx_URXD   = 0x0000
	URXD_CHARRDY = 15
	URXD_ERR     = 14
	URXD_OVRRUN  = 13
	URXD_FRMERR  = 12
	URXD_BRK     = 11
	URXD_PRERR   = 10
	URXD_RX_DATA = 0

	UARTx_UTXD   = 0x0040
	UTXD_TX_DATA = 0

	UARTx_UCR1    = 0x0080
	UCR1_ADEN     = 15
	UCR1_ADBR     = 14
	UCR1_TRDYEN   = 13
	UCR1_IDEN     = 12
	UCR1_ICD      = 10
	UCR1_RRDYEN   = 9
	UCR1_RXDMAEN  = 8
	UCR1_IREN     = 7
	UCR1_TXMPTYEN = 6
	UCR1_RTSDEN   = 5
	UCR1_SNDBRK   = 4
	UCR1_TXDMAEN  = 3
	UCR1_ATDMAEN  = 2
	UCR1_DOZE     = 1
	UCR1_UARTEN   = 0

	UARTx_UCR2 = 0x0084
	UCR2_ESCI  = 15
	UCR2_IRTS  = 14
	UCR2_CTSC  = 13
	UCR2_CTS   = 12
	UCR2_ESCEN = 11
	UCR2_RTEC  = 9
	UCR2_PREN  = 8
	UCR2_PROE  = 7
	UCR2_STPB  = 6
	UCR2_WS    = 5
	UCR2_RTSEN = 4
	UCR2_ATEN  = 3
	UCR2_TXEN  = 2
	UCR2_RXEN  = 1
	UCR2_SRST  = 0

	UARTx_UCR3     = 0x0088
	UCR3_DPEC      = 14
	UCR3_DTREN     = 13
	UCR3_PARERREN  = 12
	UCR3_FRAERREN  = 11
	UCR3_DSR       = 10
	UCR3_DCD       = 9
	UCR3_RI        = 8
	UCR3_ADNIMP    = 7
	UCR3_RXDSEN    = 6
	UCR3_AIRINTEN  = 5
	UCR3_AWAKEN    = 4
	UCR3_DTRDEN    = 3
	UCR3_RXDMUXSEL = 2
	UCR3_INVT      = 1
	UCR3_ACIEN     = 0

	UARTx_UCR4 = 0x008c
	UCR4_CTSTL = 10

	UARTx_UFCR  = 0x0090
	UFCR_TXTL   = 10
	UFCR_RFDIV  = 7
	UFCR_DCEDTE = 6
	UFCR_RXTL   = 0

	UARTx_USR2 = 0x0098
	USR2_RDR   = 0

	UARTx_UESC = 0x009c
	UARTx_UTIM = 0x00a0
	UARTx_UBIR = 0x00a4
	UARTx_UBMR = 0x00a8

	UARTx_UTS  = 0x00b4
	UTS_TXFULL = 4
)

// UART represents a serial port instance.
type UART struct {
	// Controller index
	Index int
	// Base register
	Base uint32
	// Clock gate register
	CCGR uint32
	// Clock gate
	CG int
	// Clock retrieval function
	Clock func() uint32
	// port speed
	Baudrate uint32
	// DTE mode
	DTE bool
	// hardware flow control
	Flow bool

	// control registers
	urxd uint32
	utxd uint32
	ucr1 uint32
	ucr2 uint32
	ucr3 uint32
	ucr4 uint32
	ufcr uint32
	usr2 uint32
	uesc uint32
	utim uint32
	ubir uint32
	ubmr uint32
	uts  uint32
}

// Init initializes and enables the UART for RS-232 mode,
// p3605, 55.13.1 Programming the UART in RS-232 mode, IMX6ULLRM.
func (hw *UART) Init() {
	if hw.Base == 0 || hw.CCGR == 0 || hw.Clock == nil {
		panic("invalid UART controller instance")
	}

	if hw.Baudrate == 0 {
		hw.Baudrate = UART_DEFAULT_BAUDRATE
	}

	hw.urxd = hw.Base + UARTx_URXD
	hw.utxd = hw.Base + UARTx_UTXD
	hw.ucr1 = hw.Base + UARTx_UCR1
	hw.ucr2 = hw.Base + UARTx_UCR2
	hw.ucr3 = hw.Base + UARTx_UCR3
	hw.ucr4 = hw.Base + UARTx_UCR4
	hw.ufcr = hw.Base + UARTx_UFCR
	hw.usr2 = hw.Base + UARTx_USR2
	hw.uesc = hw.Base + UARTx_UESC
	hw.utim = hw.Base + UARTx_UTIM
	hw.ubir = hw.Base + UARTx_UBIR
	hw.ubmr = hw.Base + UARTx_UBMR
	hw.uts = hw.Base + UARTx_UTS

	// enable clock
	reg.SetN(hw.CCGR, hw.CG, 0b11, 0b11)

	hw.setup()
}

func (hw *UART) txFull() bool {
	return reg.Get(hw.uts, UTS_TXFULL, 1) == 1
}

func (hw *UART) rxReady() bool {
	return reg.Get(hw.usr2, USR2_RDR, 1) == 1
}

func (hw *UART) setup() {
	// disable UART
	reg.Write(hw.ucr1, 0)
	reg.Write(hw.ucr2, 0)

	// wait for software reset deassertion
	reg.Wait(hw.ucr2, UCR2_SRST, 1, 1)

	var ucr3 uint32
	// Data Set Ready
	bits.Set(&ucr3, UCR3_DSR)
	// Data Carrier Detect
	bits.Set(&ucr3, UCR3_DCD)
	// Ring Indicator
	bits.Set(&ucr3, UCR3_RI)
	// Disable autobaud detection
	bits.Set(&ucr3, UCR3_ADNIMP)
	// UART is in MUXED mode
	bits.Set(&ucr3, UCR3_RXDMUXSEL)
	// set UCR3
	reg.Write(hw.ucr3, ucr3)

	// set escape character
	reg.Write(hw.uesc, ESC)
	// reset escape timer
	reg.Write(hw.utim, 0)

	var ufcr uint32
	// Divide input clock by 2
	bits.SetN(&ufcr, UFCR_RFDIV, 0b111, 0b100)
	// TxFIFO has 2 or fewer characters
	bits.SetN(&ufcr, UFCR_TXTL, 0b111111, 2)
	// RxFIFO has 1 character
	bits.SetN(&ufcr, UFCR_RXTL, 0b111111, 1)

	if hw.DTE {
		bits.Set(&ufcr, UFCR_DCEDTE)
	}

	// set UFCR
	reg.Write(hw.ufcr, ufcr)

	// p3592, 55.5 Binary Rate Multiplier (BRM), IMX6ULLRM
	//
	//              ref_clk_freq
	// baudrate = -----------------
	//                   UBMR + 1
	//             16 * ----------
	//                   UBIR + 1
	//
	// ref_clk_freq = module_clock

	// multiply to match UFCR_RFDIV divider value
	ubmr := hw.Clock() / (2 * hw.Baudrate)
	// neutralize denominator
	reg.Write(hw.ubir, 15)
	// set UBMR
	reg.Write(hw.ubmr, ubmr)

	var ucr2 uint32
	// 8-bit transmit and receive character length
	bits.Set(&ucr2, UCR2_WS)
	// Enable the transmitter
	bits.Set(&ucr2, UCR2_TXEN)
	// Enable the receiver
	bits.Set(&ucr2, UCR2_RXEN)
	// Software reset
	bits.Set(&ucr2, UCR2_SRST)

	if hw.Flow {
		// Receiver controls CTS
		bits.Set(&ucr2, UCR2_CTSC)

		// 16 characters in the RxFIFO as the maximum value leads to
		// overflow even with hardware flow control in place.
		reg.SetN(hw.ucr4, UCR4_CTSTL, 0b111111, 16)
	} else {
		// Ignore the RTS pin
		bits.Set(&ucr2, UCR2_IRTS)

		// 32 characters in the RxFIFO (maximum)
		reg.SetN(hw.ucr4, UCR4_CTSTL, 0b111111, 32)
	}

	// set UCR2
	reg.Write(hw.ucr2, ucr2)
	// Enable the UART
	reg.Set(hw.ucr1, UCR1_UARTEN)
}

// Enable enables the UART, this is only required after an explicit disable
// (see Disable()) as initialized interfaces (see Init()) are enabled by default.
func (hw *UART) Enable() {
	reg.Set(hw.ucr1, UCR1_UARTEN)
}

// Disable disables the UART.
func (hw *UART) Disable() {
	reg.Clear(hw.ucr1, UCR1_UARTEN)
}

// Tx transmits a single character to the serial port.
func (hw *UART) Tx(c byte) {
	for hw.txFull() {
		// wait for TX FIFO to have room for a character
	}
	reg.Write(hw.utxd, uint32(c))
}

// Rx receives a single character from the serial port.
func (hw *UART) Rx() (c byte, valid bool) {
	if !hw.rxReady() {
		return
	}

	urxd := reg.Read(hw.urxd)

	if bits.Get(&urxd, URXD_PRERR, 0b11111) != 0 {
		return
	}

	return byte(bits.Get(&urxd, URXD_RX_DATA, 0xff)), true
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
