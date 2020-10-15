// NXP i.MX6 I2C driver
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package imx6

import (
	"errors"
	"sync"
	"time"

	"github.com/f-secure-foundry/tamago/internal/reg"
)

// I2C registers
const (
	// p1462, 31.7 I2C Memory Map/Register Definition, IMX6ULLRM

	// i.MX 6UltraLite (G0, G1, G2, G3, G4)
	// i.MX 6ULL (Y0, Y1, Y2)
	// i.MX 6ULZ (Z0)
	I2C1_BASE = 0x021a0000
	I2C2_BASE = 0x021a4000

	// i.MX 6UltraLite (G1, G2, G3, G4)
	// i.MX 6ULL (Y1, Y2)
	I2C3_BASE = 0x021a8000
	I2C4_BASE = 0x021f8000

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

// I2C represents a I2C port instance.
type I2C struct {
	sync.Mutex

	// controller index
	n int
	// clock gate
	cg int

	// control registers
	iadr uint32
	ifdr uint32
	i2cr uint32
	i2sr uint32
	i2dr uint32

	// Timeout for I2C operations
	Timeout time.Duration
}

// I2C1 instance
var I2C1 = &I2C{n: 1}

// I2C2 instance
var I2C2 = &I2C{n: 2}

// Init initializes the I2C controller instance. At this time only master mode
// is supported by this driver.
func (hw *I2C) Init() {
	var base uint32

	hw.Lock()

	switch hw.n {
	case 1:
		base = I2C1_BASE
		hw.cg = CCGR2_CG3
	case 2:
		base = I2C2_BASE
		hw.cg = CCGR2_CG4
	case 3:
		base = I2C3_BASE
		hw.cg = CCGR2_CG5
	case 4:
		base = I2C4_BASE
		hw.cg = CCGR6_CG12
	default:
		panic("invalid I2C controller instance")
	}

	hw.iadr = base + I2Cx_IADR
	hw.ifdr = base + I2Cx_IFDR
	hw.i2cr = base + I2Cx_I2CR
	hw.i2sr = base + I2Cx_I2SR
	hw.i2dr = base + I2Cx_I2DR

	hw.Timeout = 1 * time.Millisecond

	hw.enable()

	hw.Unlock()
}

// getRootClock returns the PERCLK_CLK_ROOT frequency,
// (p629, Figure 18-2. Clock Tree - Part 1, IMX6ULLRM).
func (hw *I2C) getRootClock() uint32 {
	var freq uint32

	if reg.Get(CCM_CSCMR1, CSCMR1_PERCLK_SEL, 1) == 1 {
		freq = OSC_FREQ
	} else {
		// IPG_CLK_ROOT derived from AHB_CLK_ROOT which is 132 MHz
		ipg_podf := reg.Get(CCM_CBCDR, CBCDR_IPG_PODF, 0b11)
		freq = 132000000 / (ipg_podf + 1)
	}

	podf := reg.Get(CCM_CSCMR1, CSCMR1_PERCLK_PODF, 0x3f)

	return freq / (podf + 1)
}

// p1452, 31.5.1 Initialization sequence, IMX6ULLRM
func (hw *I2C) enable() {
	var register uint32

	if hw.n == 4 {
		register = CCM_CCGR2
	} else {
		register = CCM_CCGR6
	}

	reg.SetN(register, hw.cg, 0b11, 0b11)

	// Set SCL frequency
	// 66 MHz / 768 = 85 kbps
	// TODO: allow Init() to set the baudrate.
	reg.Write16(hw.ifdr, 0x16)

	reg.Set16(hw.i2cr, I2CR_IEN)
}

// Read reads a sequence of bytes from a slave device
// (p167, 16.4.2 Programming the I2C controller for I2C Read, IMX6FG).
//
// The return data buffer always matches the requested size, otherwise an error
// is returned.
//
// The address length (`alen`) parameter should be set greater then 0 for
// ordinary I2C reads (`SLAVE W|ADDR|SLAVE R|DATA`), equal to 0 when not
// sending a register address (`SLAVE W|SLAVE R|DATA`) and less than 0 only to
// send a slave read (`SLAVE R|DATA`).
func (hw *I2C) Read(slave uint8, addr uint32, alen int, size int) (buf []byte, err error) {
	hw.Lock()
	defer hw.Unlock()

	if err = hw.start(false); err != nil {
		return
	}

	if err = hw.txAddress(slave, addr, alen); err != nil {
		return
	}

	if err = hw.start(true); err != nil {
		return
	}

	// send slave address with R/W bit set
	a := byte((slave << 1) | 1)

	if err = hw.tx([]byte{a}); err != nil {
		return
	}

	buf = make([]byte, size)

	if err = hw.rx(buf); err != nil {
		return
	}

	err = hw.stop()

	return
}

// Write writes a sequence of bytes to a slave device
// (p170, 16.4.4 Programming the I2C controller for I2C Write, IMX6FG)
//
// Set greater then 0 for ordinary I2C write (`SLAVE W|ADDR|DATA`),
// set equal then 0 to not send register address (`SLAVE W|DATA`),
// alen less then 0 is invalid.

// The address length (`alen`) parameter should be set greater then 0 for
// ordinary I2C writes (`SLAVE W|ADDR|DATA`), equal to 0 when not sending a
// register address (`SLAVE W|DATA`), values less than 0 are not valid.
func (hw *I2C) Write(buf []byte, slave uint8, addr uint32, alen int) (err error) {
	if alen < 0 {
		return errors.New("invalid address length")
	}

	hw.Lock()
	defer hw.Unlock()

	if err = hw.start(false); err != nil {
		return
	}

	if err = hw.txAddress(slave, addr, alen); err != nil {
		return
	}

	if err = hw.tx(buf); err != nil {
		return
	}

	err = hw.stop()

	return
}

func (hw *I2C) txAddress(slave uint8, addr uint32, alen int) (err error) {
	if slave > 0x7f {
		return errors.New("invalid slave address")
	}

	if alen >= 0 {
		// send slave slave address with R/W bit unset
		a := byte(slave << 1)

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

	// set read from slave bit
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

		if i == size-1 {
			if err = hw.stop(); err != nil {
				return
			}
		}

		if i == size-2 {
			reg.Set16(hw.i2cr, I2CR_TXAK)
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
	if repeat == false {
		// wait for bus to be free
		if !reg.WaitFor16(hw.Timeout, hw.i2sr, I2SR_IBB, 1, 0) {
			return errors.New("timeout waiting bus to be free")
		}

		// enable master mode, generates START signal
		reg.Set16(hw.i2cr, I2CR_MSTA)
	} else {
		reg.Set16(hw.i2cr, I2CR_RSTA)
	}

	// wait for bus to be busy
	if !reg.WaitFor16(hw.Timeout, hw.i2sr, I2SR_IBB, 1, 1) {
		return errors.New("timeout waiting bus to be busy")
	}

	if repeat == false {
		// set Master Transmit mode
		reg.Set16(hw.i2cr, I2CR_MTX)
	}

	return
}

func (hw *I2C) stop() (err error) {
	reg.Clear16(hw.i2cr, I2CR_MSTA)
	reg.Clear16(hw.i2cr, I2CR_MTX)

	// wait for bus to be free
	if !reg.WaitFor16(hw.Timeout, hw.i2sr, I2SR_IBB, 1, 0) {
		err = errors.New("timeout waiting for free bus")
	}

	return
}
