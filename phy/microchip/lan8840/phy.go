// LAN8840 Ethernet PHY support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package lan8840 implements a driver for the Microchip LAN8840 Gigabit
// Ethernet Transceiver adopting the following specifications:
//   - DS00004727A (09-30-22)
package lan8840

import (
	"errors"
	"fmt"
	"time"

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/phy"
)

// PHY registers
const (
	BASIC_CONTROL   = 0x00
	BASIC_STATUS    = 0x01
	PHY_ID_1        = 0x02
	PHY_ID_2        = 0x03
	PHY_ANEG_ADV    = 0x04
	PHY_ANEG_LPA    = 0x05
	PHY_1000_CTRL   = 0x09
	PHY_1000_STATUS = 0x0a

	CTRL_RESET        = 15
	CTRL_ANEG_ENABLE  = 12
	CTRL_ANEG_RESTART = 9

	STATUS_ANEG_COMPLETE = 5
	STATUS_LINK          = 2

	ANEG_LPA_1000_HALF = 10
	ANEG_LPA_1000_FULL = 11
	ANEG_ADV_1000_FULL = 9
	ANEG_ADV_1000_HALF = 8
	ANEG_ADV_100_FULL  = 8
	ANEG_ADV_100_HALF  = 7
	ANEG_ADV_10_FULL   = 6
	ANEG_ADV_10_HALF   = 5
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

// PHY represents a LAN8840 PHY port.
type PHY struct {
	// Timeout for PHY operations
	Timeout time.Duration

	miim  phy.MIIM
	pa    int
	speed int
}

// Init initializes a LAN8840 PHY instance for register access.
func (hw *PHY) Init(addr int, miim phy.MIIM) (err error) {
	if miim == nil || addr < 0 || addr > 0x07 {
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

		if !bits.Get16(&control, CTRL_RESET) {
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
	var basic, local, partner uint16

	if basic, err = hw.link(); err != nil {
		return
	}

	hw.speed = 0

	status.Link = bits.Get16(&basic, STATUS_LINK)
	status.AutoNegotiationComplete = bits.Get16(&basic, STATUS_ANEG_COMPLETE)

	if !status.Link || !status.AutoNegotiationComplete {
		return
	}

	if local, err = hw.read(PHY_1000_CTRL); err != nil {
		return status, fmt.Errorf("could not read 1000BASE-T control register, %w", err)
	}

	if partner, err = hw.read(PHY_1000_STATUS); err != nil {
		return status, fmt.Errorf("could not read 1000BASE-T status register, %w", err)
	}

	switch {
	case bits.Get16(&local, ANEG_ADV_1000_FULL) && bits.Get16(&partner, ANEG_LPA_1000_FULL):
		status.Speed = 1000
		status.FullDuplex = true
	case bits.Get16(&local, ANEG_ADV_1000_HALF) && bits.Get16(&partner, ANEG_LPA_1000_HALF):
		status.Speed = 1000
	default:
		if local, err = hw.read(PHY_ANEG_ADV); err != nil {
			return status, fmt.Errorf("could not read auto-negotiation advertisement register, %w", err)
		}

		if partner, err = hw.read(PHY_ANEG_LPA); err != nil {
			return status, fmt.Errorf("could not read auto-negotiation link partner ability register, %w", err)
		}

		common := local & partner

		switch {
		case bits.Get16(&common, ANEG_ADV_100_FULL):
			status.Speed = 100
			status.FullDuplex = true
		case bits.Get16(&common, ANEG_ADV_100_HALF):
			status.Speed = 100
		case bits.Get16(&common, ANEG_ADV_10_FULL):
			status.Speed = 10
			status.FullDuplex = true
		case bits.Get16(&common, ANEG_ADV_10_HALF):
			status.Speed = 10
		default:
			return status, errors.New("PHY has no common advertised mode")
		}
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
