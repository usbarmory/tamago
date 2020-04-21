// NXP i.MX6 UART driver
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build tamago,arm

package imx6

import (
	"github.com/f-secure-foundry/tamago/imx6/internal/reg"
)

const (
	// base addresses

	// i.MX 6UltraLite (G0, G1, G2, G3, G4)
	// i.MX 6ULL (Y0, Y1, Y2)
	// i.MX 6ULZ (Z0)
	UART1_BASE uint32 = 0x02020000
	UART2_BASE        = 0x021e8000
	UART3_BASE        = 0x021ec000
	UART4_BASE        = 0x021f0000

	// i.MX 6UltraLite (G1, G2, G3, G4)
	// i.MX 6ULL (Y1, Y2)
	UART5_BASE        = 0x021f4000
	UART6_BASE        = 0x021fc000
	UART7_BASE        = 0x02018000
	UART8_BASE        = 0x02024000

	// Register definitions

	URXD uint32 = 0x0000
	UTXD        = 0x0040
	UCR1        = 0x0080
	UCR2        = 0x0084
	UCR3        = 0x0088
	UCR4        = 0x008c
	UFCR        = 0x0090
	USR1        = 0x0094
	USR2        = 0x0098
	UESC        = 0x009c
	UTIM        = 0x00a0
	UBIR        = 0x00a4
	UBMR        = 0x00a8
	UBRC        = 0x00ac
	ONEMS       = 0x00b0
	UTS         = 0x00b4
	UMCR        = 0x00b8

	// bit positions

	URXD_RX_DATA = 0
	URXD_PRERR   = 10
	URXD_BRK     = 11
	URXD_FRMERR  = 12
	URXD_OVRRUN  = 13
	URXD_ERR     = 14
	URXD_CHARRDY = 15

	UTXD_TX_DATA = 0

	UCR1_UARTEN   = 0
	UCR1_DOZE     = 1
	UCR1_ATDMAEN  = 2
	UCR1_TXDMAEN  = 3
	UCR1_SNDBRK   = 4
	UCR1_RTSDEN   = 5
	UCR1_TXMPTYEN = 6
	UCR1_IREN     = 7
	UCR1_RXDMAEN  = 8
	UCR1_RRDYEN   = 9
	UCR1_ICD      = 10
	UCR1_IDEN     = 12
	UCR1_TRDYEN   = 13
	UCR1_ADBR     = 14
	UCR1_ADEN     = 15

	UCR2_SRST  = 0
	UCR2_RXEN  = 1
	UCR2_TXEN  = 2
	UCR2_ATEN  = 3
	UCR2_RTSEN = 4
	UCR2_WS    = 5
	UCR2_STPB  = 6
	UCR2_PROE  = 7
	UCR2_PREN  = 8
	UCR2_RTEC  = 9
	UCR2_ESCEN = 11
	UCR2_CTS   = 12
	UCR2_CTSC  = 13
	UCR2_IRTS  = 14
	UCR2_ESCI  = 15

	UCR3_ACIEN     = 0
	UCR3_INVT      = 1
	UCR3_RXDMUXSEL = 2
	UCR3_DTRDEN    = 3
	UCR3_AWAKEN    = 4
	UCR3_AIRINTEN  = 5
	UCR3_RXDSEN    = 6
	UCR3_ADNIMP    = 7
	UCR3_RI        = 8
	UCR3_DCD       = 9
	UCR3_DSR       = 10
	UCR3_FRAERREN  = 11
	UCR3_PARERREN  = 12
	UCR3_DTREN     = 13
	UCR3_DPEC      = 14

	UCR4_DREN    = 0
	UCR4_OREN    = 1
	UCR4_BKEN    = 2
	UCR4_TCEN    = 3
	UCR4_LPBYP   = 4
	UCR4_IRSC    = 5
	UCR4_IDDMAEN = 6
	UCR4_WKEN    = 7
	UCR4_ENIRI   = 8
	UCR4_INVR    = 9
	UCR4_CTSTL   = 10

	UFCR_RXTL   = 0
	UFCR_DCEDTE = 6
	UFCR_RFDIV  = 7
	UFCR_TXTL   = 10

	UFCR_RXTL_MASK   = 0b111111 << UFCR_RXTL
	UFCR_DCEDTE_MASK = 1 << UFCR_DCEDTE
	UFCR_RFDIV_MASK  = 1 << UFCR_RFDIV
	UFCR_TXTL_MASK   = 1 << UFCR_TXTL

	// misc
	UTS_TXEMPTY = 6
	USR2_RDR    = 0

)

type uart struct {
	urxd uint32
	utxd uint32
	ucr1 uint32
	ucr2 uint32
	ucr3 uint32
	ucr4 uint32
	ufcr uint32
	usr1 uint32
	usr2 uint32
	uesc uint32
	utim uint32
	ubir uint32
	ubmr uint32
	ubrc uint32
	onems uint32
	uts uint32
	umcr uint32
}

func (u *uart) Init(base uint32) {
	u.urxd  = base + URXD
	u.utxd  = base + UTXD
	u.ucr1  = base + UCR1
	u.ucr2  = base + UCR2
	u.ucr3  = base + UCR3
	u.ucr4  = base + UCR4
	u.ufcr  = base + UFCR
	u.usr1  = base + USR1
	u.usr2  = base + USR2
	u.uesc  = base + UESC
	u.utim  = base + UTIM
	u.ubir  = base + UBIR
	u.ubmr  = base + UBMR
	u.ubrc  = base + UBRC
	u.onems = base + ONEMS
	u.uts   = base + UTS
	u.umcr  = base + UMCR
}

func uartclk() uint32 {
	var CCM ccm
	var podf, freq uint32

	CCM.Init(CCM_BASE)

	if (reg.Get(CCM.cscdr1, CSCDR1_UART_CLK_SEL, 0b1) == 1) {
		freq = OSC_CLK
	} else {
		freq = 480000000
	}

	podf = reg.Get(CCM.cscdr1, CSCDR1_CLK_PODF, 0b111111)

	return freq / (podf + 1)
}

func (u *uart) txEmpty() bool {
	return reg.Get(u.uts, UTS_TXEMPTY, 0b1) == 0
}

func (u *uart) rxReady() bool {
	return reg.Get(u.usr2, USR2_RDR, 0b1) == 1
}

func (u *uart) rxError() bool {
	return reg.Get(u.urxd, URXD_PRERR, 0b11111) != 0
}

//              ref_clk_freq
// baudrate = -----------------
//                   UBMR + 1
//             16 * ----------
//                   UBIR + 1

// ref_clk_freq = module_clock

func (u *uart) Setup(baudrate uint32) bool {
	var tmp uint32
	var clk uint32

	// Disable UART
	reg.Write(u.ucr1, 0)
	reg.Write(u.ucr2, 0) // 0x4027

	for (reg.Get(u.ucr2, UCR2_SRST, 0b1) == 0) {
		// wait for software reset deasserted
	}

	reg.Write(u.ucr3, 0x704 | (1 << UCR3_ADNIMP));
	reg.Write(u.ucr4, 0x8000)
	reg.Write(u.uesc, 0x2b)
	reg.Write(u.utim, 0)

	clk = uartclk() / 6

	tmp = 4 << UFCR_RFDIV
	tmp |= (2 << UFCR_TXTL) | (1 << UFCR_RXTL)
	reg.Write(u.ufcr, tmp)
	reg.Write(u.ubir, 0xf)
	tmp = clk / (2 * baudrate)
	reg.Write(u.ubmr, tmp)
	tmp = (1 << UCR2_WS) | (1 << UCR2_IRTS) | (1 << UCR2_RXEN) | (1 << UCR2_TXEN) | (1 << UCR2_SRST)
	reg.Write(u.ucr2, tmp)
	reg.Write(u.ucr1, 1 << UCR1_UARTEN)

	return true
}

// Write a single character to the selected serial port.
func (u *uart) Write(c byte) {
	// transmit data
	reg.Write(u.utxd, uint32(c))

	for u.txEmpty() {
		// wait for TX FIFO to be empty
	}
}

// Read a single character from the selected serial port.
func (u *uart) Read() (c byte, valid bool) {
	if !u.rxReady() {
		return c, false
	}

	if u.rxError() {
		return c, false
	}

	return byte(reg.Get(u.urxd, URXD_RX_DATA, 0xff)), true
}
