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
	"github.com/f-secure-foundry/tamago/internal/reg"
)

const (
	UART1_URXD uint32 = 0x02020000
	UART1_UTXD uint32 = 0x02020040
	UART1_UTS  uint32 = 0x020200b4
	UART1_USR2 uint32 = 0x02020098

	UART2_URXD uint32 = 0x021e8000
	UART2_UTXD uint32 = 0x021e8040
	UART2_UTS  uint32 = 0x021e80b4
	UART2_USR2 uint32 = 0x021e8098

	UART_URXD_PRERR   = 10
	UART_URXD_RX_DATA = 0
	UART_UTS_TXEMPTY  = 6
	UART_USR2_RDR     = 0
)

type uart struct {
	urxd uint32
	utxd uint32
	uts  uint32
	usr2 uint32
}

var UART1 = &uart{
	urxd: UART1_URXD,
	utxd: UART1_UTXD,
	uts:  UART1_UTS,
	usr2: UART1_USR2,
}

var UART2 = &uart{
	urxd: UART2_URXD,
	utxd: UART2_UTXD,
	uts:  UART2_UTS,
	usr2: UART2_USR2,
}

func (u *uart) txEmpty() bool {
	return reg.Get(u.uts, UART_UTS_TXEMPTY, 1) == 0
}

func (u *uart) rxReady() bool {
	return reg.Get(u.usr2, UART_USR2_RDR, 1) == 1
}

func (u *uart) rxError() bool {
	return reg.Get(u.urxd, UART_URXD_PRERR, 0b11111) != 0
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

	return byte(reg.Get(u.urxd, UART_URXD_RX_DATA, 0xff)), true
}
