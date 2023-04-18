// NXP I2C driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package i2c implements a driver for NXP I2C controllers adopting the
// following reference specifications:
//   - IMX6ULLRM - i.MX 6ULL Applications Processor Reference Manual - Rev 1 2017/11
//   - IMX6FG    - i.MX 6 Series Firmware Guide                      - Rev 0 2012/11
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/usbarmory/tamago.
package i2c

import (
	"errors"
	"sync"
	"time"

	"github.com/usbarmory/tamago/internal/reg"
)

// I2C registers
// (p1462, 31.7 I2C Memory Map/Register Definition, IMX6ULLRM)
const (
	// The default IFDR value corresponds to a frequency divider of 768,
	// assuming 66 MHz for PERCLK_CLK_ROOT this results in a baud rate of
	// 85 kbps (p1464, 31.7.2 I2C Frequency Divider Register (I2Cx_IFDR),
	// IMX6ULLRM).
	I2C_DEFAULT_IFDR = 0x16

	I2Cx_IADR = 0x0000
	I2Cx_IFDR = 0x0004

	I2Cx_I2CR = 0x0008
	I2CR_IEN  = 7
	I2CR_MSTA = 5
	I2CR_MTX  = 4
	I2CR_TXAK = 3
	I2CR_RSTA = 2

	I2Cx_I2SR = 0x000c
	I2SR_IBB  = 5
	I2SR_IIF  = 1
	I2SR_RXAK = 0

	I2Cx_I2DR = 0x0010
)

// Configuration constants
const (
	// Timeout is the default timeout for I2C operations.
	Timeout = 100 * time.Millisecond
)

// I2C represents an I2C port instance.
type I2C struct {
	sync.Mutex

	// Controller index
	Index int
	// Base register
	Base uint32
	// Clock gate register
	CCGR uint32
	// Clock gate
	CG int
	// Timeout for I2C operations
	Timeout time.Duration
	// Div sets the frequency divider to control the I2C clock rate
	// (p1464, 31.7.2 I2C Frequency Divider Register (I2Cx_IFDR), IMX6ULLRM).
	Div uint16

	// control registers
	iadr uint32
	ifdr uint32
	i2cr uint32
	i2sr uint32
	i2dr uint32
}

// Init initializes the I2C controller instance. At this time only master mode
// is supported by this driver.
func (hw *I2C) Init() {
	hw.Lock()
	defer hw.Unlock()

	if hw.Base == 0 || hw.CCGR == 0 {
		panic("invalid I2C controller instance")
	}

	if hw.Timeout == 0 {
		hw.Timeout = Timeout
	}

	if hw.Div == 0 {
		hw.Div = I2C_DEFAULT_IFDR
	}

	hw.iadr = hw.Base + I2Cx_IADR
	hw.ifdr = hw.Base + I2Cx_IFDR
	hw.i2cr = hw.Base + I2Cx_I2CR
	hw.i2sr = hw.Base + I2Cx_I2SR
	hw.i2dr = hw.Base + I2Cx_I2DR

	// p1452, 31.5.1 Initialization sequence, IMX6ULLRM

	// enable clock
	reg.SetN(hw.CCGR, hw.CG, 0b11, 0b11)

	// Set SCL frequency
	reg.Write16(hw.ifdr, hw.Div)

	reg.Set16(hw.i2cr, I2CR_IEN)
}

// Read reads a sequence of bytes from a target device
// (p167, 16.4.2 Programming the I2C controller for I2C Read, IMX6FG).
//
// The return data buffer always matches the requested size, otherwise an error
// is returned.
//
// The address length (`alen`) parameter should be set greater then 0 for
// ordinary I2C reads (`SLAVE W|ADDR|SLAVE R|DATA`), equal to 0 when not
// sending a register address (`SLAVE W|SLAVE R|DATA`) and less than 0 only to
// send a target read (`SLAVE R|DATA`).
func (hw *I2C) Read(target uint8, addr uint32, alen int, size int) (buf []byte, err error) {
	hw.Lock()
	defer hw.Unlock()

	if err = hw.start(false); err != nil {
		return
	}
	defer hw.stop()

	if alen > 0 {
		if err = hw.txAddress(target, addr, alen); err != nil {
			return
		}

		if err = hw.start(true); err != nil {
			return
		}
	}

	// send target address with R/W bit set
	a := byte((target << 1) | 1)

	if err = hw.tx([]byte{a}); err != nil {
		return
	}

	buf = make([]byte, size)
	err = hw.rx(buf)

	return
}

// Write writes a sequence of bytes to a target device
// (p170, 16.4.4 Programming the I2C controller for I2C Write, IMX6FG)
//
// Set greater then 0 for ordinary I2C write (`SLAVE W|ADDR|DATA`),
// set equal then 0 to not send register address (`SLAVE W|DATA`),
// alen less then 0 is invalid.
//
// The address length (`alen`) parameter should be set greater then 0 for
// ordinary I2C writes (`SLAVE W|ADDR|DATA`), equal to 0 when not sending a
// register address (`SLAVE W|DATA`), values less than 0 are not valid.
func (hw *I2C) Write(buf []byte, target uint8, addr uint32, alen int) (err error) {
	if alen < 0 {
		return errors.New("invalid address length")
	}

	hw.Lock()
	defer hw.Unlock()

	if err = hw.start(false); err != nil {
		return
	}
	defer hw.stop()

	if err = hw.txAddress(target, addr, alen); err != nil {
		return
	}

	return hw.tx(buf)
}

func (hw *I2C) txAddress(target uint8, addr uint32, alen int) (err error) {
	if target > 0x7f {
		return errors.New("invalid target address")
	}

	if alen > 4 {
		return errors.New("invalid register address length")
	}

	if alen >= 0 {
		// send target address with R/W bit unset
		a := byte(target << 1)

		if err = hw.tx([]byte{a}); err != nil {
			return
		}
	}

	// send register address
	for alen > 0 {
		alen--
		a := byte(addr >> (alen * 8) & 0xff)

		if err = hw.tx([]byte{a}); err != nil {
			return
		}
	}

	return
}

func (hw *I2C) rx(buf []byte) (err error) {
	size := len(buf)

	// set read from target bit
	reg.Clear16(hw.i2cr, I2CR_MTX)

	if size == 1 {
		reg.Set16(hw.i2cr, I2CR_TXAK)
	} else {
		reg.Clear16(hw.i2cr, I2CR_TXAK)
	}

	reg.Clear16(hw.i2sr, I2SR_IIF)
	// dummy read
	reg.Read16(hw.i2dr)

	for i := 0; i < size; i++ {
		if !reg.WaitFor16(hw.Timeout, hw.i2sr, I2SR_IIF, 1, 1) {
			return errors.New("timeout on byte reception")
		}

		if i == size-2 {
			reg.Set16(hw.i2cr, I2CR_TXAK)
		} else if i == size-1 {
			hw.stop()
		}

		buf[i] = byte(reg.Read16(hw.i2dr) & 0xff)
		reg.Clear16(hw.i2sr, I2SR_IIF)
	}

	return
}

func (hw *I2C) tx(buf []byte) (err error) {
	for i := 0; i < len(buf); i++ {
		reg.Clear16(hw.i2sr, I2SR_IIF)
		reg.Write16(hw.i2dr, uint16(buf[i]))

		if !reg.WaitFor16(hw.Timeout, hw.i2sr, I2SR_IIF, 1, 1) {
			return errors.New("timeout on byte transmission")
		}

		if reg.Get16(hw.i2sr, I2SR_RXAK, 1) == 1 {
			return errors.New("no acknowledgement received")
		}
	}

	return
}

func (hw *I2C) start(repeat bool) (err error) {
	var pos int

	if repeat == false {
		// wait for bus to be free
		if !reg.WaitFor16(hw.Timeout, hw.i2sr, I2SR_IBB, 1, 0) {
			return errors.New("timeout waiting bus to be free")
		}

		// enable master mode, generates START signal
		pos = I2CR_MSTA
	} else {
		pos = I2CR_RSTA
	}

	reg.Set16(hw.i2cr, pos)

	// wait for bus to be busy
	if !reg.WaitFor16(hw.Timeout, hw.i2sr, I2SR_IBB, 1, 1) {
		reg.Clear16(hw.i2cr, pos)
		return errors.New("timeout waiting bus to be busy")
	}

	if repeat == false {
		// set Master Transmit mode
		reg.Set16(hw.i2cr, I2CR_MTX)
	}

	return
}

func (hw *I2C) stop() {
	reg.Clear16(hw.i2cr, I2CR_MSTA)
	reg.Clear16(hw.i2cr, I2CR_MTX)
}
