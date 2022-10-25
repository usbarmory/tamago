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
	MDIO_ST       = 0b01
	MDIO_OP_READ  = 0b10
	MDIO_OP_WRITE = 0b01
	MDIO_TA       = 0b10
)

func mdio22(op int, pa int, ra int, data uint16) (pkt uint32) {
	bits.SetN(&pkt, MMFR_ST, 0b11, MDIO_ST)
	bits.SetN(&pkt, MMFR_OP, 0b11, uint32(op))
	bits.SetN(&pkt, MMFR_PA, 0x1f, uint32(pa))
	bits.SetN(&pkt, MMFR_RA, 0x1f, uint32(ra))
	bits.SetN(&pkt, MMFR_TA, 0b11, MDIO_TA)
	bits.SetN(&pkt, MMFR_DATA, 0xffff, uint32(data))

	return
}

// ReadMII reads a connected Ethernet PHY register.
func (hw *ENET) ReadMII(pa int, ra int) (data uint16) {
	reg.Set(hw.eir, EIR_MII)
	defer reg.Set(hw.eir, EIR_MII)

	reg.Write(hw.mmfr, mdio22(MDIO_OP_READ, pa, ra, 0))
	reg.Wait(hw.eir, EIR_MII, 1, 1)

	return uint16(reg.Read(hw.mmfr))
}

// WriteMII writes a connected Ethernet PHY register.
func (hw *ENET) WriteMII(pa int, ra int, data uint16) {
	reg.Set(hw.eir, EIR_MII)
	defer reg.Set(hw.eir, EIR_MII)

	reg.Write(hw.mmfr, mdio22(MDIO_OP_WRITE, pa, ra, data))
	reg.Wait(hw.eir, EIR_MII, 1, 1)
}
