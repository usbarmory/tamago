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
	CTRL_SPEED0       = 13
	CTRL_ANEG_ENABLE  = 12
	CTRL_ANEG_RESTART = 9
	CTRL_DUPLEX       = 8
	CTRL_SPEED1       = 6

	STATUS_LINK          = 2
	STATUS_ANEG_COMPLETE = 5

	ANEG_ADV_10_HALF   = 1 << 5
	ANEG_ADV_10_FULL   = 1 << 6
	ANEG_ADV_100_HALF  = 1 << 7
	ANEG_ADV_100_FULL  = 1 << 8
	ANEG_ADV_1000_HALF = 1 << 8
	ANEG_ADV_1000_FULL = 1 << 9
	ANEG_LPA_1000_HALF = 1 << 10
	ANEG_LPA_1000_FULL = 1 << 11
)

// WaitTimeout represents the timeout for PHY register writes
var WaitTimeout = 10 * time.Second

// PollInterval represents the delay between PHY register write attempts
var PollInterval = 10 * time.Millisecond

// Status represents the resolved PHY link state.
type Status struct {
	Link                    bool
	AutoNegotiationComplete bool
	Speed                   int
	FullDuplex              bool
}

// PHY represents the LAN8840 Ethernet PHY instance.
type PHY struct {
	miim  phy.MIIM
	pa    int
	speed int
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

func (hw *PHY) wait(addr int, mask uint16, value uint16) (data uint16, err error) {
	deadline := time.Now().Add(WaitTimeout)

	for {
		if data, err = hw.read(addr); err != nil {
			return
		}

		if data&mask == value {
			return
		}

		if time.Now().After(deadline) {
			return 0, fmt.Errorf("timed out waiting for PHY register %#x", addr)
		}

		time.Sleep(PollInterval)
	}
}

// Init initializes the PHY and performs [PHY.Negotiate].
func (hw *PHY) Init(addr int, miim phy.MIIM) (err error) {
	if miim == nil {
		return errors.New("invalid PHY instance")
	}

	hw.miim = miim
	hw.pa = addr
	hw.speed = 0

	// software reset
	if err = hw.write(BASIC_CONTROL, (1 << CTRL_RESET)); err != nil {
		return
	}

	if _, err = hw.wait(BASIC_CONTROL, 1<<CTRL_RESET, 0); err != nil {
		return fmt.Errorf("could not reset PHY, %v", err)
	}

	if err = hw.Negotiate(); err != nil {
		return
	}

	statusMask := uint16((1 << STATUS_LINK) | (1 << STATUS_ANEG_COMPLETE))

	if _, err = hw.wait(BASIC_STATUS, statusMask, statusMask); err != nil {
		return fmt.Errorf("auto-negotiation status error, %w", err)
	}

	var status Status
	if status, err = hw.Status(); err != nil {
		return
	}

	switch {
	case status.Speed == 10:
		return errors.New("unsupported 10 Mbps management link")
	case !status.FullDuplex:
		return errors.New("unsupported half-duplex management link")
	case status.Speed != 100 && status.Speed != 1000:
		return errors.New("management PHY has no common advertised mode")
	}

	return
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

	return basic&(1<<STATUS_LINK) != 0, nil
}

// Status returns link, auto-negotiation, speed, and duplex state.
func (hw *PHY) Status() (status Status, err error) {
	var basic uint16

	if basic, err = hw.link(); err != nil {
		return
	}

	hw.speed = 0

	status.Link = basic&(1<<STATUS_LINK) != 0
	status.AutoNegotiationComplete = basic&(1<<STATUS_ANEG_COMPLETE) != 0

	if !status.Link || !status.AutoNegotiationComplete {
		return
	}

	var local, partner uint16

	if local, err = hw.read(PHY_1000_CTRL); err != nil {
		return status, fmt.Errorf("could not read 1000BASE-T control register, %v", err)
	}

	if partner, err = hw.read(PHY_1000_STATUS); err != nil {
		return status, fmt.Errorf("could not read 1000BASE-T status register, %v", err)
	}

	switch {
	case local&ANEG_ADV_1000_FULL != 0 && partner&ANEG_LPA_1000_FULL != 0:
		status.Speed = 1000
		status.FullDuplex = true
	case local&ANEG_ADV_1000_HALF != 0 && partner&ANEG_LPA_1000_HALF != 0:
		status.Speed = 1000
	default:
		if local, err = hw.read(PHY_ANEG_ADV); err != nil {
			return status, fmt.Errorf("could not read auto-negotiation advertisement register, %v", err)
		}

		if partner, err = hw.read(PHY_ANEG_LPA); err != nil {
			return status, fmt.Errorf("could not read auto-negotiation link partner ability register, %v", err)
		}

		common := local & partner

		switch {
		case common&ANEG_ADV_100_FULL != 0:
			status.Speed = 100
			status.FullDuplex = true
		case common&ANEG_ADV_100_HALF != 0:
			status.Speed = 100
		case common&ANEG_ADV_10_FULL != 0:
			status.Speed = 10
			status.FullDuplex = true
		case common&ANEG_ADV_10_HALF != 0:
			status.Speed = 10
		default:
			return status, errors.New("management PHY has no common advertised mode")
		}
	}

	hw.speed = status.Speed

	return
}

// Address returns the PHY address passed at [PHY.Init].
func (hw *PHY) Address() int {
	return hw.pa
}

// Speed returns the PHY auto-negotiated speed.
func (hw *PHY) Speed() int {
	return hw.speed
}
