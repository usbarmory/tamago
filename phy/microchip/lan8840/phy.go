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
	PHY_CTRL        = 0x00
	PHY_STATUS      = 0x01
	PHY_ANEG_ADV    = 0x04
	PHY_ANEG_LPA    = 0x05
	PHY_1000_CTRL   = 0x09
	PHY_1000_STATUS = 0x0a

	CTRL_RESET        = 15
	CTRL_SPEED0       = 13
	CTRL_ANEG         = 12
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

// PHY represents the LAN8840 Ethernet PHY instance.
type PHY struct {
	// Address represents the PHY address and must be set before [Init]
	Address int

	// MIIM represents the MIIM interface and must be set before [Init]
	MIIM    phy.MIIM

	speed int
}

func (phy *PHY) wait(addr int, mask uint16, value uint16) (data uint16, err error) {
	deadline := time.Now().Add(WaitTimeout)

	for {
		if data, err = phy.MIIM.ReadPHYRegister(phy.Address, addr); err != nil {
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
func (phy *PHY) Init() (err error) {
	if phy.MIIM == nil {
		return errors.New("invalid PHY instance")
	}

	phy.speed = 0

	// software reset
	if err = phy.MIIM.WritePHYRegister(phy.Address, PHY_CTRL, (1 << CTRL_RESET)); err != nil {
		return
	}

	control, err := phy.wait(PHY_CTRL, 1<<CTRL_RESET, 0)

	if err != nil {
		return fmt.Errorf("could not reset PHY, %v", err)
	}

	// enable and restart auto-negotiation
	control |= (1 << CTRL_ANEG) | (1 << CTRL_ANEG_RESTART)

	if err = phy.MIIM.WritePHYRegister(phy.Address, PHY_CTRL, control); err != nil {
		return
	}

	statusMask := uint16((1 << STATUS_LINK) | (1 << STATUS_ANEG_COMPLETE))

	if _, err = phy.wait(PHY_STATUS, statusMask, statusMask); err != nil {
		return fmt.Errorf("auto-negotiation status error, %w", err)
	}

	return phy.Negotiate()
}

// Negotiate refreshes the PHY auto-negotiation status, currently only 100/1000
// Mbps Auto-Negotiation (Full-duplex) is supported.
func (phy *PHY) Negotiate() (err error) {
	var local, partner uint16

	phy.speed = 0

	if phy.MIIM == nil {
		return errors.New("invalid PHY instance")
	}

	if local, err = phy.MIIM.ReadPHYRegister(phy.Address, PHY_1000_CTRL); err != nil {
		return fmt.Errorf("could not read 1000BASE-T control register, %v", err)
	}

	if partner, err = phy.MIIM.ReadPHYRegister(phy.Address, PHY_1000_STATUS); err != nil {
		return fmt.Errorf("could not read 1000BASE-T status register, %v", err)
	}

	if local&ANEG_ADV_1000_FULL != 0 && partner&ANEG_LPA_1000_FULL != 0 {
		phy.speed = 1000
		return
	}

	if local&ANEG_ADV_1000_HALF != 0 && partner&ANEG_LPA_1000_HALF != 0 {
		return fmt.Errorf("unsupported half-duplex management link")
	}

	if local, err = phy.MIIM.ReadPHYRegister(phy.Address, PHY_ANEG_ADV); err != nil {
		return fmt.Errorf("could not read auto-negotiation advertisement register, %v", err)
	}

	if partner, err = phy.MIIM.ReadPHYRegister(phy.Address, PHY_ANEG_LPA); err != nil {
		return fmt.Errorf("could not read auto-negotiation link partner ability register, %v", err)
	}

	common := local & partner

	switch {
	case common&ANEG_ADV_100_FULL != 0:
		phy.speed = 100
	case common&ANEG_ADV_100_HALF != 0:
		return fmt.Errorf("unsupported half-duplex management link")
	case common&(ANEG_ADV_10_FULL|ANEG_ADV_10_HALF) != 0:
		return fmt.Errorf("unsupported 10 Mbps management link")
	default:
		return fmt.Errorf("management PHY has no common advertised mode")
	}

	return
}

// Speed returns the PHY auto-negotiated speed.
func (phy *PHY) Speed() int {
	return phy.speed
}
