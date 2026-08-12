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
	"errors"
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

	// TWI represents the TWI mode configuration
	TWI *TWI
	// USART represents the USART mode configuration
	USART *USART
}

func (hw *FLEXCOM) mode(n int) {
	reg.SetN(hw.Base+FLEX_MR, MR_OPMODE, MR_OPMODE_MASK, uint32(n))
}

// InitUSART initializes and enables a FLEXCOM controller instance in USART
// mode, this is mutually exclusive with other modes.
func (hw *FLEXCOM) InitUSART() {
	hw.Lock()

	if hw.Base == 0 {
		panic("invalid FLEXCOM controller instance")
	}

	if hw.USART == nil {
		hw.USART = &USART{}
	}

	hw.mode(MR_OPMODE_USART)

	hw.USART.cr = hw.Base + FLEX_USART_OFFSET + FLEX_US_CR
	hw.USART.mr = hw.Base + FLEX_USART_OFFSET + FLEX_US_MR
	hw.USART.ier = hw.Base + FLEX_USART_OFFSET + FLEX_US_IER
	hw.USART.idr = hw.Base + FLEX_USART_OFFSET + FLEX_US_IDR
	hw.USART.csr = hw.Base + FLEX_USART_OFFSET + FLEX_US_CSR
	hw.USART.rhr = hw.Base + FLEX_USART_OFFSET + FLEX_US_RHR
	hw.USART.thr = hw.Base + FLEX_USART_OFFSET + FLEX_US_THR
	hw.USART.brgr = hw.Base + FLEX_USART_OFFSET + FLEX_US_BRGR
	hw.USART.fmr = hw.Base + FLEX_USART_OFFSET + FLEX_US_FMR

	hw.USART.init()

	// no defer as goos.Hwinit1 might call us from system stack
	hw.Unlock()
}

// InitTWI initializes and enables a FLEXCOM controller in TWI initiator mode,
// this is mutually exclusive with other modes.
func (hw *FLEXCOM) InitTWI() (err error) {
	hw.Lock()
	defer hw.Unlock()

	if hw.Base == 0 {
		return errors.New("invalid FLEXCOM controller instance")
	}

	if hw.TWI == nil {
		hw.TWI = &TWI{}
	}

	hw.mode(MR_OPMODE_TWI)
	hw.TWI.Base = hw.Base

	return hw.TWI.init()
}
