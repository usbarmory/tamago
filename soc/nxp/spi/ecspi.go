// NXP Enhanced Configurable SPI (ECSPI) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package spi implements a driver for NXP SPI controllers adopting the
// following reference specifications:
//   - IMX6ULLRM - i.MX 6ULL Applications Processor Reference Manual - Rev 1 2017/11
//
// This package is only meant to be used with `GOOS=tamago` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package spi

import (
	"encoding/binary"
	"errors"
	"sync"
	"time"

	"github.com/usbarmory/tamago/internal/reg"
)

// SPI registers
// (p805, 20.7 ECSPI memory map, IMX6ULLRM)
const (
	// The default divider values correspond to a resulting SPI clock of 1
	// MHz.
	ECSPI_DEFAULT_PRE_DIVIDER  = 14
	ECSPI_DEFAULT_POST_DIVIDER = 2

	ECSPIx_RXDATA = 0x00
	ECSPIx_TXDATA = 0x04

	ECSPIx_CONREG         = 0x08
	CONREG_BURST_LENGTH   = 20
	CONREG_CHANNEL_SELECT = 18
	CONREG_PRE_DIVIDER    = 12
	CONREG_POST_DIVIDER   = 8
	CONREG_CHANNEL_MODE   = 4
	CONREG_XCH            = 2
	CONREG_EN             = 0

	ECSPIx_CONFIGREG = 0x0c
	CONFIGREG_SS_POL = 12
	CONFIGREG_SS_CTL = 8

	ECSPIx_STATREG = 0x18
	STATREG_TC     = 7
	STATREG_RR     = 3
	STATREG_TF     = 2

	ECSPIx_MSGDATA = 0x40
)

// Timeout is the default timeout for SPI operations.
const Timeout = 100 * time.Millisecond

// ECSPI represents an ECSPI port instance.
type ECSPI struct {
	sync.Mutex

	// Controller index
	Index int
	// Base register
	Base uint32
	// Clock gate register
	CCGR uint32
	// Clock gate
	CG int

	// PreDiv sets the 1st stage SPI frequency divider
	PreDiv int
	// PostDiv sets the 2nd stage SPI frequency divider
	PostDiv int
	// Timeout for I2C operations
	Timeout time.Duration

	// Channel defines the Chip Select assertion on [Transfer]
	Channel int

	// control registers
	rxdata  uint32
	txdata  uint32
	conreg  uint32
	statreg uint32
}

// Init initializes an ECSPI controller instance in master mode 0 (CPOL=0
// CPHA=0).
func (hw *ECSPI) Init() {
	hw.Lock()
	defer hw.Unlock()

	if hw.Base == 0 || hw.CCGR == 0 {
		panic("invalid ECSPI controller instance")
	}

	if hw.Timeout == 0 {
		hw.Timeout = Timeout
	}

	hw.rxdata = hw.Base + ECSPIx_RXDATA
	hw.txdata = hw.Base + ECSPIx_TXDATA
	hw.conreg = hw.Base + ECSPIx_CONREG
	hw.statreg = hw.Base + ECSPIx_STATREG

	// p804, 20.5 Initialization, IMX6ULLRM

	// reset
	reg.Clear(hw.conreg, CONREG_EN)

	// set 8-bit burst length
	reg.SetN(hw.conreg, CONREG_BURST_LENGTH, 0xfff, 7)

	// enable clock
	reg.SetN(hw.CCGR, hw.CG, 0b11, 0b11)

	// set master mode for all channels
	reg.SetN(hw.conreg, CONREG_CHANNEL_MODE, 0xf, 0xf)

	if hw.PreDiv == 0 && hw.PostDiv == 0 {
		hw.PreDiv = ECSPI_DEFAULT_PRE_DIVIDER
		hw.PostDiv = ECSPI_DEFAULT_POST_DIVIDER
	}

	// set SPI frequency
	reg.SetN(hw.conreg, CONREG_PRE_DIVIDER, 0xf, uint32(hw.PreDiv))
	reg.SetN(hw.conreg, CONREG_POST_DIVIDER, 0xf, uint32(hw.PostDiv))

	// out of reset
	reg.Set(hw.conreg, CONREG_EN)
}

// Transfer performs a full-duplex SPI exchange in-place. On return, buf
// contains the received bytes.
func (hw *ECSPI) Transfer(buf []byte) (err error) {
	hw.Lock()
	defer hw.Unlock()

	if hw.Channel < 0 || hw.Channel > 3 {
		return errors.New("invalid channel")
	}

	reg.SetN(hw.conreg, CONREG_CHANNEL_SELECT, 0b11, uint32(hw.Channel))
	reg.SetN(hw.conreg, CONREG_BURST_LENGTH, 0xfff, uint32(len(buf)*8)-1)

	var b []byte

	for i := 0; i < len(buf); i += 4 {
		if reg.Get(hw.statreg, STATREG_TF) {
			return errors.New("transmit FIFO full")
		}

		b = buf[i:]

		if n := len(b); n < 4 {
			b = append(make([]byte, 4-n), b...)
		}

		reg.Write(hw.txdata, binary.BigEndian.Uint32(b))
	}

	// initiate exchange
	reg.Set(hw.conreg, CONREG_XCH)

	// wait for exchange completion
	if !reg.WaitFor(hw.Timeout, hw.conreg, CONREG_XCH, 1, 0) {
		return errors.New("exchange timeout")
	}

	// wait for transfer completion
	if !reg.WaitFor(hw.Timeout, hw.statreg, STATREG_TC, 1, 1) {
		return errors.New("transfer timout")
	}

	// read response
	for i := 0; i < len(buf); i += 4 {
		if !reg.Get(hw.statreg, STATREG_RR) {
			break
		}

		b = buf[i:]

		if n := len(b); n < 4 {
			r := make([]byte, 4)
			binary.BigEndian.PutUint32(r, reg.Read(hw.rxdata))
			copy(b, r)
		} else {
			binary.BigEndian.PutUint32(b, reg.Read(hw.rxdata))
		}
	}

	// clear transfer completion
	reg.Set(hw.statreg, STATREG_TC)

	return
}
