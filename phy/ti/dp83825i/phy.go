// DP83825I Ethernet PHY support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package dp83825i implements a driver for the Texas Instruments DP83825I
// Ethernet Transceiver adopting the following specifications:
//   - SNLS638C – DECEMBER 2018 – REVISED APRIL 2026
package dp83825i

import (
	"errors"
	"fmt"

	"github.com/usbarmory/tamago/phy"
)

// PHY registers
const (
	DP_CTRL     = 0x00
	CTRL_RESET  = 15
	CTRL_SPEED  = 13
	CTRL_ANEG   = 12
	CTRL_DUPLEX = 8

	DP_REGCR = 0xd
	DP_ADDAR = 0xe

	DP_RCSR      = 0x17
	RCSR_RMII_CS = 7
	RCSR_RX_BUF  = 0
)

// PHY represents the DP83825I Ethernet PHY instance.
type PHY struct {
	// Address represents the PHY address and must be set before [Init]
	Address int

	// MIIM represents the MIIM interface and must be set before [Init]
	MIIM    phy.MIIM

	speed int
}

// Init initializes the PHY and performs [PHY.Negotiate].
func (phy *PHY) Init() (err error) {
	if phy.MIIM == nil {
		return errors.New("invalid PHY instance")
	}

	phy.speed = 0

	// software reset
	if err = phy.MIIM.WritePHYRegister(phy.Address, DP_CTRL, (1 << CTRL_RESET)); err != nil {
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

	// 100 Mbps, Auto-Negotiation, Full-duplex
	ctrl := (1<<CTRL_SPEED)|(1<<CTRL_ANEG)|(1<<CTRL_DUPLEX)

	if err = phy.MIIM.WritePHYRegister(phy.Address, DP_CTRL, uint16(ctrl)); err != nil {
		return fmt.Errorf("could not configure auto-negotiation, %v", err)
	}

	// 50MHz RMII Reference Clock Select, 2 bit tolerance Receive Elasticity Buffer Size
	rcsr := (1<<RCSR_RMII_CS)|(1<<RCSR_RX_BUF)

	if err = phy.MIIM.WritePHYRegister(phy.Address, DP_RCSR, uint16(rcsr)); err != nil {
		return fmt.Errorf("could not select reference clock, %v", err)
	}

	phy.speed = 100

	return
}

// Speed returns the PHY auto-negotiated speed.
func (phy *PHY) Speed() int {
	return phy.speed
}
