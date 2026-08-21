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
	"sync"

	"github.com/usbarmory/tamago/internal/reg"
)

// FLEXCOM registers
const (
	// Peripheral clock is 250 MHz
	// p501, 3.47.3 CLOCKING AND RESET, Microchip DS00005048E
	PERIPHERAL_CLK = 250e6

	FLEX_MR         = 0x00
	MR_OPMODE       = 0
	MR_OPMODE_MASK  = 0x3
	MR_OPMODE_USART = 1
	MR_OPMODE_TWI   = 3
)

// FLEXCOM represents a Flexible Serial Communication controller instance.
type FLEXCOM struct {
	sync.Mutex

	// Controller index
	Index int
	// Base register
	Base uint32
	// Interrupt ID
	IRQ int
}

func (hw *FLEXCOM) mode(n int) {
	reg.SetN(hw.Base+FLEX_MR, MR_OPMODE, MR_OPMODE_MASK, uint32(n))
}
