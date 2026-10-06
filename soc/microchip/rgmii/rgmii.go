// Microchip RGMII Ethernet device support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package rgmii implements speed configuration for Microchip LAN969x RGMII
// interfaces adopting the following specifications:
//   - Microchip - LAN9694/LAN9696/LAN9698 Datasheet - DS00005048E (02-27-25)
//
// This package is only meant to be used with `GOOS=tamago` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
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

	SPEED_SEL      = 20
	SPEED_SEL_MASK = 0b111
	SPEED_10M      = 0
	SPEED_100M     = 1
	SPEED_1G       = 2

	MAC_TX_RST = 4
	MAC_RX_RST = 0

	MAC_ENA_CFG = 0x24
	RX_ENA      = 4
	TX_ENA      = 0
)

// HSIOWRAP RGMII_CFG fields
const (
	TX_CLK_CFG      = 2
	TX_CLK_CFG_MASK = 0b111
	TX_CLK_10M      = 3
	TX_CLK_100M     = 2
	TX_CLK_1G       = 1

	RGMII_TX_RST = 1
	RGMII_RX_RST = 0
)

// RGMII represents an RGMII interface instance.
type RGMII struct {
	sync.Mutex

	// DEVRGMII base register
	Base uint32
	// HSIOWRAP RGMII_CFG register
	ClockConfig uint32
}

// SetSpeed configures the MAC and transmit clock for 10, 100 or 1000 Mbps
// operation, frames in transit may be discarded.
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
		txClock = TX_CLK_10M
		macSpeed = SPEED_10M
	case 100:
		txClock = TX_CLK_100M
		macSpeed = SPEED_100M
	case 1000:
		txClock = TX_CLK_1G
		macSpeed = SPEED_1G
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
	reg.SetN(hw.ClockConfig, TX_CLK_CFG, TX_CLK_CFG_MASK, txClock)
	reg.SetN(hw.Base+DEV_RST_CTRL, SPEED_SEL, SPEED_SEL_MASK, macSpeed)

	// release clock domains
	reg.Clear(hw.ClockConfig, RGMII_TX_RST)
	reg.Clear(hw.ClockConfig, RGMII_RX_RST)
	reg.Clear(hw.Base+DEV_RST_CTRL, MAC_TX_RST)
	reg.Clear(hw.Base+DEV_RST_CTRL, MAC_RX_RST)
	reg.Write(hw.Base+MAC_ENA_CFG, enable)

	return
}
