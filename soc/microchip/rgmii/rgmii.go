// Microchip RGMII Ethernet device support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package rgmii implements speed changes for Microchip LAN969x RGMII MACs
// under the LAN9694/LAN9696/LAN9698 Datasheet, DS00005048E (02-27-25).
package rgmii

import (
	"errors"
	"fmt"
	"sync"

	"github.com/usbarmory/tamago/internal/reg"
)

// DEVRGMII registers
const (
	DEV_RST_CTRL = 0x00
	SPEED_SEL    = 20
	MAC_TX_RST   = 4
	MAC_RX_RST   = 0

	MAC_ENA_CFG = 0x24
	RX_ENA      = 4
	TX_ENA      = 0
)

// HSIOWRAP RGMII_CFG fields
const (
	TX_CLK_CFG   = 2
	RGMII_TX_RST = 1
	RGMII_RX_RST = 0
)

// RGMII represents one RGMII MAC and its clock configuration.
type RGMII struct {
	sync.Mutex

	// Base is the DEVRGMII register base address.
	Base uint32
	// ClockConfig is the associated HSIOWRAP RGMII_CFG register address.
	ClockConfig uint32
}

// SetSpeed matches MAC and RGMII clocks to 10, 100, or 1000 Mbps without
// changing PHY, pin, DLL, forwarding, or MAC enable configuration. The MAC
// must already be initialized. Callers must exclude packet I/O; a change may
// discard frames. Invalid instances and speeds leave hardware unchanged.
func (hw *RGMII) SetSpeed(speed int) (err error) {
	hw.Lock()
	defer hw.Unlock()

	if hw.Base == 0 || hw.ClockConfig == 0 {
		return errors.New("invalid RGMII instance")
	}

	var txClock uint32
	var macSpeed uint32

	switch speed {
	case 10:
		txClock = 3
		macSpeed = 0
	case 100:
		txClock = 2
		macSpeed = 1
	case 1000:
		txClock = 1
		macSpeed = 2
	default:
		return fmt.Errorf("invalid RGMII speed (%d)", speed)
	}

	// stop MAC paths
	enable := reg.Read(hw.Base + MAC_ENA_CFG)
	reg.Clear(hw.Base+MAC_ENA_CFG, RX_ENA)
	reg.Clear(hw.Base+MAC_ENA_CFG, TX_ENA)
	reg.Set(hw.Base+DEV_RST_CTRL, MAC_TX_RST)
	reg.Set(hw.Base+DEV_RST_CTRL, MAC_RX_RST)
	reg.Set(hw.ClockConfig, RGMII_TX_RST)
	reg.Set(hw.ClockConfig, RGMII_RX_RST)

	// match clock selectors
	reg.SetN(hw.ClockConfig, TX_CLK_CFG, 0b111, txClock)
	reg.SetN(hw.Base+DEV_RST_CTRL, SPEED_SEL, 0b111, macSpeed)

	// release clock domains
	reg.Clear(hw.ClockConfig, RGMII_TX_RST)
	reg.Clear(hw.ClockConfig, RGMII_RX_RST)
	reg.Clear(hw.Base+DEV_RST_CTRL, MAC_TX_RST)
	reg.Clear(hw.Base+DEV_RST_CTRL, MAC_RX_RST)
	reg.Write(hw.Base+MAC_ENA_CFG, enable)

	return
}
