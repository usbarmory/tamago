// NXP 10/100-Mbps Ethernet MAC (ENET)
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package enet

import (
	"errors"
	"time"

	"github.com/usbarmory/tamago/internal/mdio"
	"github.com/usbarmory/tamago/internal/reg"
)

// Timeout is the default timeout for MIIM operations.
const Timeout = 1 * time.Second

// MDIO22 transmits an MII frame (IEEE 802.3-2008 Clause 22) to a connected
// Ethernet PHY, the transacted frame is returned.
func (hw *ENET) MDIO22(op, pa, ra int, data uint16) (frame uint32, err error) {
	reg.Set(hw.eir, IRQ_MII)
	defer reg.Set(hw.eir, IRQ_MII)

	frame = mdio.Frame(mdio.ST, uint32(op), uint32(pa), uint32(ra), mdio.TA, data)
	reg.Write(hw.mmfr, frame)

	if !reg.WaitFor(hw.Timeout, hw.eir, IRQ_MII, 1, 1) {
		return 0, errors.New("command timeout")
	}

	return reg.Read(hw.mmfr), nil
}

// MDIO45 transmits an MII frame (IEEE 802.3-2008 Clause 45) to a connected
// Ethernet PHY, the transacted frame is returned.
func (hw *ENET) MDIO45(op, prtad, devad int, data uint16) (frame uint32, err error) {
	reg.Set(hw.eir, IRQ_MII)
	defer reg.Set(hw.eir, IRQ_MII)

	frame = mdio.Frame(mdio.ST_45, uint32(op), uint32(prtad), uint32(devad), mdio.TA_45, data)
	reg.Write(hw.mmfr, frame)

	if !reg.WaitFor(hw.Timeout, hw.eir, IRQ_MII, 1, 1) {
		return 0, errors.New("command timeout")
	}

	return reg.Read(hw.mmfr), nil
}

// ReadPHYRegister reads a standard management register of a connected Ethernet
// PHY (IEEE 802.3-2008 Clause 22).
func (hw *ENET) ReadPHYRegister(pa int, ra int) (data uint16, err error) {
	frame, err := hw.MDIO22(mdio.OP_READ, pa, ra, 0)
	return uint16(frame), err
}

// WritePHYRegister writes a standard management register of a connected
// Ethernet PHY (IEEE 802.3-2008 Clause 22).
func (hw *ENET) WritePHYRegister(pa int, ra int, data uint16) (err error) {
	_, err = hw.MDIO22(mdio.OP_WRITE, pa, ra, data)
	return
}
