// NXP 10/100-Mbps Ethernet MAC (ENET)
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package enet

import (
	"github.com/usbarmory/tamago/internal/mdio"
	"github.com/usbarmory/tamago/internal/reg"
)

// MDIO22 transmits an MII frame (IEEE 802.3-2008 Clause 22) to a connected
// Ethernet PHY, the transacted frame is returned.
func (hw *ENET) MDIO22(op, pa, ra int, data uint16) (frame uint32) {
	reg.Set(hw.eir, IRQ_MII)
	defer reg.Set(hw.eir, IRQ_MII)

	frame = mdio.Frame(mdio.ST, uint32(op), uint32(pa), uint32(ra), mdio.TA, data)
	reg.Write(hw.mmfr, frame)

	reg.Wait(hw.eir, IRQ_MII, 1, 1)
	return reg.Read(hw.mmfr)
}

// MDIO45 transmits an MII frame (IEEE 802.3-2008 Clause 45) to a connected
// Ethernet PHY, the transacted frame is returned.
func (hw *ENET) MDIO45(op, prtad, devad int, data uint16) (frame uint32) {
	reg.Set(hw.eir, IRQ_MII)
	defer reg.Set(hw.eir, IRQ_MII)

	frame = mdio.Frame(mdio.ST_45, uint32(op), uint32(prtad), uint32(devad), mdio.TA_45, data)
	reg.Write(hw.mmfr, frame)

	reg.Wait(hw.eir, IRQ_MII, 1, 1)
	return reg.Read(hw.mmfr)
}

// ReadPHYRegister reads a standard management register of a connected Ethernet
// PHY (IEE 802.3-2008 Clause 22).
func (hw *ENET) ReadPHYRegister(pa int, ra int) (data uint16) {
	return uint16(hw.MDIO22(mdio.OP_READ, pa, ra, 0))
}

// WritePHYRegister writes a standard management register of a connected
// Ethernet PHY (IEE 802.3-2008 Clause 22).
func (hw *ENET) WritePHYRegister(pa int, ra int, data uint16) {
	hw.MDIO22(mdio.OP_WRITE, pa, ra, data)
}
