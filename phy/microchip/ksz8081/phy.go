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
	miim  phy.MIIM
	pa    int
	speed int
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
	if err = hw.miim.WritePHYRegister(hw.pa, KSZ_CTRL, (1 << CTRL_RESET)); err != nil {
		return
	}

	return hw.Negotiate()
}

// Negotiate refreshes the PHY auto-negotiation status, currently only 100 Mbps
// Auto-Negotiation (Full-duplex) is supported.
func (hw *PHY) Negotiate() (err error) {
	hw.speed = 0

	if hw.miim == nil {
		return errors.New("invalid PHY instance")
	}

	// HP Auto MDI/MDI-X mode, RMII 50MHz, LEDs: Activity/Link
	ctrl := (1 << CTRL2_HP_MDIX) | (1 << CTRL2_RMII) | (1 << CTRL2_LED)

	if err = hw.miim.WritePHYRegister(hw.pa, KSZ_PHYCTRL2, uint16(ctrl)); err != nil {
		return
	}

	// 100 Mbps, Full-duplex
	ctrl = (1 << CTRL_SPEED) | (1 << CTRL_DUPLEX)

	if err = hw.miim.WritePHYRegister(hw.pa, KSZ_CTRL, uint16(ctrl)); err != nil {
		return
	}

	hw.speed = 100

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

// EnableInterrupts enables all available transceiver interrupts.
func (hw *PHY) EnableInterrupts() (err error) {
	if hw.miim == nil {
		return errors.New("invalid PHY instance")
	}

	return hw.miim.WritePHYRegister(hw.pa, KSZ_INT, 0xff00)
}
