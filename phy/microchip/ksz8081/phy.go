// KSZ8081RNB Ethernet PHY support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package ksz8081 implements a driver for the Microchip KSZ8081RNB Ethernet
// Transceiver adopting the following specifications:
//   - DS000002202C (01-09-18)
package ksz8081

import (
	"errors"

	"github.com/usbarmory/tamago/phy"
)

// PHY registers
const (
	KSZ_CTRL    = 0x00
	CTRL_RESET  = 15
	CTRL_SPEED  = 13
	CTRL_DUPLEX = 8

	KSZ_INT = 0x1b

	KSZ_PHYCTRL2  = 0x1f
	CTRL2_HP_MDIX = 15
	CTRL2_RMII    = 7
	CTRL2_LED     = 4
)

// PHY represents the KSZ8081 Ethernet PHY instance.
type PHY struct {
	// Address represents the PHY address and must be set before [Init]
	Address int

	// MIIM represents the MIIM interface and must be set before [Init]
	MIIM phy.MIIM

	speed int
}

// Init initializes the PHY and performs [PHY.Negotiate].
func (phy *PHY) Init() (err error) {
	if phy.MIIM == nil {
		return errors.New("invalid PHY instance")
	}

	phy.speed = 0

	// software reset
	if err = phy.MIIM.WritePHYRegister(phy.Address, KSZ_CTRL, (1 << CTRL_RESET)); err != nil {
		return
	}

	return phy.Negotiate()
}

// Negotiate refreshes the PHY auto-negotiation status, currently only 100 Mbps
// Auto-Negotiation (Full-duplex) is supported.
func (phy *PHY) Negotiate() (err error) {
	phy.speed = 0

	if phy.MIIM == nil {
		return errors.New("invalid PHY instance")
	}

	// HP Auto MDI/MDI-X mode, RMII 50MHz, LEDs: Activity/Link
	ctrl := (1 << CTRL2_HP_MDIX) | (1 << CTRL2_RMII) | (1 << CTRL2_LED)

	if phy.MIIM.WritePHYRegister(phy.Address, KSZ_PHYCTRL2, uint16(ctrl)); err != nil {
		return
	}

	// 100 Mbps, Full-duplex
	ctrl = (1 << CTRL_SPEED) | (1 << CTRL_DUPLEX)

	if phy.MIIM.WritePHYRegister(phy.Address, KSZ_CTRL, uint16(ctrl)); err != nil {
		return
	}

	phy.speed = 100

	return
}

// Speed returns the PHY auto-negotiated speed.
func (phy *PHY) Speed() int {
	return phy.speed
}

// EnableInterrupts enables all available transceiver interrupts.
func (phy *PHY) EnableInterrupts() (err error) {
	if phy.MIIM == nil {
		return errors.New("invalid PHY instance")
	}

	return phy.MIIM.WritePHYRegister(phy.Address, KSZ_INT, 0xff00)
}
