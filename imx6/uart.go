// NXP i.MX6 UART driver
// https://github.com/inversepath/tamago
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
	"unsafe"

	"github.com/inversepath/tamago/imx6/internal/reg"
)

const UART1_URXD uint32 = 0x02020000
const UART1_UTXD uint32 = 0x02020040
const UART1_UTS uint32 = 0x020200b4
const UART1_USR2 uint32 = 0x02020098

const UART2_URXD uint32 = 0x021e8000
const UART2_UTXD uint32 = 0x021e8040
const UART2_UTS uint32 = 0x021e80b4
const UART2_USR2 uint32 = 0x021e8098

const UART_URXD_PRERR = 10
const UART_URXD_RX_DATA = 0
const UART_UTS_TXEMPTY = 6
const UART_USR2_RDR = 0

type uart struct {
	urxd *uint32
	utxd *byte
	uts  *uint32
	usr2 *uint32
}

var UART1 = &uart{
	urxd: (*uint32)(unsafe.Pointer(uintptr(UART1_URXD))),
	utxd: (*byte)(unsafe.Pointer(uintptr(UART1_UTXD))),
	uts:  (*uint32)(unsafe.Pointer(uintptr(UART1_UTS))),
	usr2: (*uint32)(unsafe.Pointer(uintptr(UART1_USR2))),
}

var UART2 = &uart{
	urxd: (*uint32)(unsafe.Pointer(uintptr(UART2_URXD))),
	utxd: (*byte)(unsafe.Pointer(uintptr(UART2_UTXD))),
	uts:  (*uint32)(unsafe.Pointer(uintptr(UART2_UTS))),
	usr2: (*uint32)(unsafe.Pointer(uintptr(UART2_USR2))),
}

func (u *uart) txEmpty() bool {
	return reg.Get(u.uts, UART_UTS_TXEMPTY, 0b1) == 0
}

func (u *uart) rxReady() bool {
	return reg.Get(u.usr2, UART_USR2_RDR, 0b1) == 1
}

func (u *uart) rxError() bool {
	return reg.Get(u.urxd, UART_URXD_PRERR, 0b11111) != 0
}

// Write a single character to the selected serial port.
func (u *uart) Write(c byte) {
	// transmit data
	*(u.utxd) = c

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
