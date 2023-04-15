// NXP 10/100-Mbps Ethernet MAC (ENET)
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package enet

import (
	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
)

const (
	// IEEE 802.3-2008 Clause 22
	MDIO_ST       = 0b01
	MDIO_OP_READ  = 0b10
	MDIO_OP_WRITE = 0b01
	MDIO_TA       = 0b10

	// IEEE 802.3-2008 Clause 45
	MDIO_45_ST          = 0b00
	MDIO_45_OP_ADDR     = 0b00
	MDIO_45_OP_WRITE    = 0b01
	MDIO_45_OP_READ_INC = 0b10
	MDIO_45_OP_READ     = 0b11
	MDIO_45_TA          = 0b10
)

func mdio(st, op, pa, ra, ta uint32, data uint16) (frame uint32) {
	bits.SetN(&frame, MMFR_ST, 0b11, st)
	bits.SetN(&frame, MMFR_OP, 0b11, op)
	bits.SetN(&frame, MMFR_PA, 0x1f, pa)
	bits.SetN(&frame, MMFR_RA, 0x1f, ra)
	bits.SetN(&frame, MMFR_TA, 0b11, ta)
	bits.SetN(&frame, MMFR_DATA, 0xffff, uint32(data))

	return
}

// MDIO22 transmits an MII frame (IEEE 802.3-2008 Clause 22) to a connected
// Ethernet PHY, the transacted frame is returned.
func (hw *ENET) MDIO22(op, pa, ra int, data uint16) (frame uint32) {
	reg.Set(hw.eir, IRQ_MII)
	defer reg.Set(hw.eir, IRQ_MII)

	frame = mdio(MDIO_ST, uint32(op), uint32(pa), uint32(ra), MDIO_TA, data)
	reg.Write(hw.mmfr, frame)

	reg.Wait(hw.eir, IRQ_MII, 1, 1)
	return reg.Read(hw.mmfr)
}

// MDIO45 transmits an MII frame (IEEE 802.3-2008 Clause 45) to a connected
// Ethernet PHY, the transacted frame is returned.
func (hw *ENET) MDIO45(op, prtad, devad int, data uint16) (frame uint32) {
	reg.Set(hw.eir, IRQ_MII)
	defer reg.Set(hw.eir, IRQ_MII)

	frame = mdio(MDIO_45_ST, uint32(op), uint32(prtad), uint32(devad), MDIO_45_TA, data)
	reg.Write(hw.mmfr, frame)

	reg.Wait(hw.eir, IRQ_MII, 1, 1)
	return reg.Read(hw.mmfr)
}

// ReadPHYRegister reads a standard management register of a connected Ethernet
// PHY (IEE 802.3-2008 Clause 22).
func (hw *ENET) ReadPHYRegister(pa int, ra int) (data uint16) {
	return uint16(hw.MDIO22(MDIO_OP_READ, pa, ra, 0))
}

// WritePHYRegister writes a standard management register of a connected
// Ethernet PHY (IEE 802.3-2008 Clause 22).
func (hw *ENET) WritePHYRegister(pa int, ra int, data uint16) {
	hw.MDIO22(MDIO_OP_WRITE, pa, ra, data)
}
