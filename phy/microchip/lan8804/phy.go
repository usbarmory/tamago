// LAN8804/LAN8814 Ethernet PHY support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package lan8804 implements a driver for the Microchip LAN8804/LAN8814
// Gigabit Ethernet PHYs adopting the following reference specifications:
//   - Microchip - LAN8804D/LAN8814D GPHY Register Definitions - DS00004286D
//   - Microchip - LAN8804D Datasheet - DS00003591K
package lan8804

import (
	"errors"
	"fmt"
	"time"

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/phy"
)

// PHY registers
const (
	BASIC_CONTROL  = 0x00
	BASIC_STATUS   = 0x01
	PHY_ID_1       = 0x02
	PHY_ID_2       = 0x03
	EP_ACCESS_CTRL = 0x16
	EP_ACCESS_DATA = 0x17
	CONTROL        = 0x1f

	CTRL_RESET        = 15
	CTRL_ANEG_ENABLE  = 12
	CTRL_ANEG_RESTART = 9

	STATUS_LINK          = 2
	STATUS_ANEG_COMPLETE = 5

	EP_FUNCTION = 14
	EP_INDEX    = 0

	EP_ADDRESS = 0
	EP_DATA    = 1

	CONTROL_SPEED_1000 = 6
	CONTROL_SPEED_100  = 5
	CONTROL_SPEED_10   = 4
	CONTROL_DUPLEX     = 3
)

// Timeout is the default timeout for PHY operations.
const Timeout = 100 * time.Millisecond

// Status represents the resolved PHY link state.
type Status struct {
	Link                    bool
	AutoNegotiationComplete bool
	Speed                   int
	FullDuplex              bool
}

// PHY represents a LAN8804/LAN8814 PHY port.
type PHY struct {
	// Timeout for PHY operations
	Timeout time.Duration

	miim  phy.MIIM
	pa    int
	speed int
}

// Init initializes a LAN8804/LAN8814 PHY instance for register access.
func (hw *PHY) Init(addr int, miim phy.MIIM) (err error) {
	if miim == nil || addr < 0 || addr > 0x1f {
		return errors.New("invalid PHY instance")
	}

	hw.miim = miim
	hw.pa = addr
	hw.speed = 0

	if hw.Timeout == 0 {
		hw.Timeout = Timeout
	}

	return
}

func (hw *PHY) read(address int) (data uint16, err error) {
	if hw.miim == nil {
		return 0, errors.New("invalid PHY instance")
	}

	return hw.miim.ReadPHYRegister(hw.pa, address)
}

func (hw *PHY) write(address int, data uint16) (err error) {
	if hw.miim == nil {
		return errors.New("invalid PHY instance")
	}

	return hw.miim.WritePHYRegister(hw.pa, address, data)
}

func (hw *PHY) selectExtended(page int, address uint16) (err error) {
	if page < 0 || page > 0x1f {
		return errors.New("invalid extended page")
	}

	ctrl := EP_ADDRESS<<EP_FUNCTION | page<<EP_INDEX

	if err = hw.write(EP_ACCESS_CTRL, uint16(ctrl)); err != nil {
		return
	}

	if err = hw.write(EP_ACCESS_DATA, address); err != nil {
		return
	}

	ctrl = EP_DATA<<EP_FUNCTION | page<<EP_INDEX

	return hw.write(EP_ACCESS_CTRL, uint16(ctrl))
}

// ReadExtendedRegister reads a vendor-specific Extended Page register.
func (hw *PHY) ReadExtendedRegister(page int, address uint16) (data uint16, err error) {
	if err = hw.selectExtended(page, address); err != nil {
		return
	}

	return hw.read(EP_ACCESS_DATA)
}

// WriteExtendedRegister writes a vendor-specific Extended Page register.
func (hw *PHY) WriteExtendedRegister(page int, address uint16, data uint16) (err error) {
	if err = hw.selectExtended(page, address); err != nil {
		return
	}

	return hw.write(EP_ACCESS_DATA, data)
}

// Reset performs a software hard reset and waits for its completion.
func (hw *PHY) Reset() (err error) {
	var control uint16

	if err = hw.write(BASIC_CONTROL, 1<<CTRL_RESET); err != nil {
		return
	}

	deadline := time.Now().Add(hw.Timeout)

	for {
		if control, err = hw.read(BASIC_CONTROL); err != nil {
			return
		}

		if bits.Get16(&control, CTRL_RESET) {
			return
		}

		if time.Now().After(deadline) {
			return errors.New("reset timeout")
		}

		time.Sleep(time.Millisecond)
	}
}

// Negotiate enables and restarts auto-negotiation.
func (hw *PHY) Negotiate() (err error) {
	var control uint16

	hw.speed = 0

	if control, err = hw.read(BASIC_CONTROL); err != nil {
		return
	}

	control |= (1 << CTRL_ANEG_ENABLE) | (1 << CTRL_ANEG_RESTART)

	return hw.write(BASIC_CONTROL, control)
}

// Identifier returns the PHY identifier registers as one 32-bit value.
func (hw *PHY) Identifier() (id uint32, err error) {
	var high, low uint16

	if high, err = hw.read(PHY_ID_1); err != nil {
		return
	}

	if low, err = hw.read(PHY_ID_2); err != nil {
		return
	}

	return uint32(high)<<16 | uint32(low), nil
}

func (hw *PHY) link() (basic uint16, err error) {
	// STATUS_LINK is latch-low; the first read clears stale link state.
	if _, err = hw.read(BASIC_STATUS); err != nil {
		return
	}

	return hw.read(BASIC_STATUS)
}

// Link reports whether the PHY link is up. The status register is read twice
// because its link bit is latched low.
func (hw *PHY) Link() (up bool, err error) {
	var basic uint16

	if basic, err = hw.link(); err != nil {
		return
	}

	return bits.Get16(&basic, STATUS_LINK), nil
}

// Status returns link, auto-negotiation, speed, and duplex state.
func (hw *PHY) Status() (status Status, err error) {
	var basic, control uint16

	if basic, err = hw.link(); err != nil {
		return
	}

	hw.speed = 0

	status.Link = bits.Get16(&basic, STATUS_LINK)
	status.AutoNegotiationComplete = bits.Get16(&basic, STATUS_ANEG_COMPLETE)

	if !status.Link {
		return
	}

	if control, err = hw.read(CONTROL); err != nil {
		return
	}

	status.FullDuplex = bits.Get16(&control, CONTROL_DUPLEX)

	switch {
	case bits.Get16(&control, CONTROL_SPEED_1000):
		status.Speed = 1000
	case bits.Get16(&control, CONTROL_SPEED_100):
		status.Speed = 100
	case bits.Get16(&control, CONTROL_SPEED_10):
		status.Speed = 10
	default:
		err = fmt.Errorf("invalid speed status %#x", control)
	}

	hw.speed = status.Speed

	return
}

// Address returns the PHY address passed at [PHY.Init].
func (hw *PHY) Address() int {
	return hw.pa
}

// Speed returns the last resolved PHY link speed.
func (hw *PHY) Speed() int {
	return hw.speed
}
