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
	"errors"
	"fmt"
	"sync"
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
	sync.Mutex

	// Controller index
	Index int
	// Base register
	Base uint32
	// Timeout for MIIM operations
	Timeout time.Duration
}

// Init initializes and enables an MIIM controller instance.
func (hw *MIIM) Init() (err error) {
	hw.Lock()
	defer hw.Unlock()

	if hw.Base == 0 {
		return errors.New("invalid MIIM controller instance")
	}

	if hw.Timeout == 0 {
		hw.Timeout = Timeout
	}

	return nil
}

func (hw *MIIM) mdio(phyad, regad, wrdata, op uint32, read bool) (rddata uint16, err error) {
	if !reg.WaitFor(hw.Timeout, hw.Base+MII_STATUS, STATUS_PENDING_OP, 1, 0) {
		return 0, errors.New("command FIFO timeout")
	}

	var cmd uint32
	bits.Set(&cmd, CMD_VLD)
	bits.SetN(&cmd, CMD_PHYAD, 0x1f, phyad)
	bits.SetN(&cmd, CMD_REGAD, 0x1f, regad)
	bits.SetN(&cmd, CMD_WRDATA, 0xffff, wrdata)
	bits.SetN(&cmd, CMD_OPR, 0b11, op)
	reg.Write(hw.Base+MII_CMD, cmd)

	if !read {
		if !reg.WaitFor(hw.Timeout, hw.Base+MII_STATUS, STATUS_PENDING_WR, 1, 0) {
			return 0, errors.New("write timeout")
		}

		return
	}

	if !reg.WaitFor(hw.Timeout, hw.Base+MII_STATUS, STATUS_BUSY, 1, 0) {
		return 0, errors.New("read timeout")
	}

	data := reg.Read(hw.Base + MII_DATA)

	if data>>DATA_SUCCESS&0b11 != 0 {
		return 0, errors.New("PHY access failed")
	}

	return uint16(data >> DATA_RDDATA), nil
}

// MDIO22 transmits an MII frame (IEEE 802.3-2008 Clause 22) to a connected
// Ethernet PHY and returns read data when applicable.
func (hw *MIIM) MDIO22(op, pa, ra int, data uint16) (rddata uint16, err error) {
	hw.Lock()
	defer hw.Unlock()

	if hw.Base == 0 {
		return 0, errors.New("invalid MIIM controller instance")
	}

	reg.SetN(hw.Base+MII_CFG, CFG_ST_CFG_FIELD, 0b11, ST_CLAUSE_22)
	read := op == mdio.OP_READ

	if rddata, err = hw.mdio(uint32(pa), uint32(ra), uint32(data), uint32(op), read); err != nil {
		return 0, fmt.Errorf("clause 22 PHY %d register %d: %w", pa, ra, err)
	}

	return
}

// MDIO45 transmits an MII frame (IEEE 802.3-2008 Clause 45) to a connected
// Ethernet PHY and returns read data when applicable.
func (hw *MIIM) MDIO45(op, prtad, devad int, data uint16) (rddata uint16, err error) {
	hw.Lock()
	defer hw.Unlock()

	if hw.Base == 0 {
		return 0, errors.New("invalid MIIM controller instance")
	}

	reg.SetN(hw.Base+MII_CFG, CFG_ST_CFG_FIELD, 0b11, ST_CLAUSE_45)
	read := op == mdio.OP_READ_INC || op == mdio.OP_READ_45

	if rddata, err = hw.mdio(uint32(prtad), uint32(devad), uint32(data), uint32(op), read); err != nil {
		return 0, fmt.Errorf("clause 45 port %d device %d: %w", prtad, devad, err)
	}

	return
}

// ReadPHYRegister reads a standard management register of a connected Ethernet
// PHY (IEEE 802.3-2008 Clause 22).
func (hw *MIIM) ReadPHYRegister(pa int, ra int) (data uint16, err error) {
	return hw.MDIO22(mdio.OP_READ, pa, ra, 0)
}

// WritePHYRegister writes a standard management register of a connected
// Ethernet PHY (IEEE 802.3-2008 Clause 22).
func (hw *MIIM) WritePHYRegister(pa int, ra int, data uint16) (err error) {
	_, err = hw.MDIO22(mdio.OP_WRITE, pa, ra, data)
	return
}
