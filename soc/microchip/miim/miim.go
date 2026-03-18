// Microchip MII Management Controller (MIIM)
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package miim implements a driver for Microchip MII Management Controllers
// (MIIM) adopting the following reference specifications:
//   - Microchip - LAN9694/LAN9696/LAN9698 Datasheet - DS00005048E (02-27-25)
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm64` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package miim

import (
	"time"

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/mdio"
	"github.com/usbarmory/tamago/internal/reg"
)

// MIIM registers
const (
	MII_STATUS        = 0x00
	STATUS_BUSY       = 3
	STATUS_PENDING_OP = 2
	STATUS_PENDING_RD = 1
	STATUS_PENDING_WR = 0

	MII_CMD    = 0x08
	CMD_VLD    = 31
	CMD_PHYAD  = 25
	CMD_REGAD  = 20
	CMD_WRDATA = 4
	CMD_OPR    = 1

	MII_DATA     = 0x0c
	DATA_SUCCESS = 16
	DATA_RDDATA  = 0

	MII_CFG          = 0x10
	CFG_ST_CFG_FIELD = 9
	ST_CLAUSE_22     = 0b01
	ST_CLAUSE_45     = 0b00
)

// Timeout is the default timeout for MIIM operations.
const Timeout = 1 * time.Second

// MIIM represents a MII Management Controller instance.
type MIIM struct {
	// Controller index
	Index int
	// Base register
	Base uint32
	// Timeout for MIIM operations
	Timeout time.Duration
}

// Init initializes and enables an MIIM controller instance.
func (hw *MIIM) Init() {
	if hw.Base == 0 {
		panic("invalid MIIM controller instance")
	}

	if hw.Timeout == 0 {
		hw.Timeout = Timeout
	}
}

func (hw *MIIM) mdio(phyad, regad, wrdata, op uint32) (rddata uint16) {
	var cmd uint32

	bits.Set(&cmd, CMD_VLD)
	bits.SetN(&cmd, CMD_PHYAD, 0x1f, phyad)
	bits.SetN(&cmd, CMD_REGAD, 0x1f, regad)
	bits.SetN(&cmd, CMD_WRDATA, 0xffff, wrdata)
	bits.SetN(&cmd, CMD_OPR, 0b11, op)

	reg.WaitFor(Timeout, hw.Base+MII_STATUS, STATUS_PENDING_OP, 1, 0)
	reg.Write(hw.Base+MII_CMD, cmd)

	if op == mdio.OP_WRITE {
		return
	}

	reg.WaitFor(Timeout, hw.Base+MII_STATUS, STATUS_BUSY, 1, 0)
	data := reg.GetN(hw.Base+MII_DATA, DATA_RDDATA, 0xffff)

	return uint16(data)
}

// MDIO22 transmits an MII frame (IEEE 802.3-2008 Clause 22) to a connected
// Ethernet PHY, the return data is returned on write operations.
func (hw *MIIM) MDIO22(op, pa, ra int, data uint16) (rddata uint16) {
	reg.SetN(hw.Base+MII_CFG, CFG_ST_CFG_FIELD, 0b11, ST_CLAUSE_22)
	return hw.mdio(uint32(pa), uint32(ra), uint32(data), uint32(op))
}

// MDIO45 transmits an MII frame (IEEE 802.3-2008 Clause 45) to a connected
// Ethernet PHY, the return data is returned on write operations.
func (hw *MIIM) MDIO45(op, prtad, devad int, data uint16) (rddata uint16) {
	reg.SetN(hw.Base+MII_CFG, CFG_ST_CFG_FIELD, 0b11, ST_CLAUSE_45)
	return hw.mdio(uint32(prtad), uint32(devad), uint32(data), uint32(op))
}

// ReadPHYRegister reads a standard management register of a connected Ethernet
// PHY (IEE 802.3-2008 Clause 22).
func (hw *MIIM) ReadPHYRegister(pa int, ra int) (data uint16) {
	return hw.MDIO22(mdio.OP_READ, pa, ra, 0)
}

// WritePHYRegister writes a standard management register of a connected
// Ethernet PHY (IEE 802.3-2008 Clause 22).
func (hw *MIIM) WritePHYRegister(pa int, ra int, data uint16) {
	hw.MDIO22(mdio.OP_WRITE, pa, ra, data)
}
