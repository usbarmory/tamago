// Management Data Input/Output (MDIO)
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package mdio

import (
	"github.com/usbarmory/tamago/bits"
)

const (
	// IEEE 802.3-2008 Clause 22
	ST       = 0b01
	OP_READ  = 0b10
	OP_WRITE = 0b01
	TA       = 0b10

	// IEEE 802.3-2008 Clause 45
	ST_45       = 0b00
	OP_ADDR     = 0b00
	OP_READ_INC = 0b10
	TA_45       = 0b10

	// Management Frame Fields
	MMFR_ST   = 30
	MMFR_OP   = 28
	MMFR_PA   = 23
	MMFR_RA   = 18
	MMFR_TA   = 16
	MMFR_DATA = 0
)

// Frame builds an MDIO frame.
func Frame(st, op, pa, ra, ta uint32, data uint16) (frame uint32) {
	bits.SetN(&frame, MMFR_ST, 0b11, st)
	bits.SetN(&frame, MMFR_OP, 0b11, op)
	bits.SetN(&frame, MMFR_PA, 0x1f, pa)
	bits.SetN(&frame, MMFR_RA, 0x1f, ra)
	bits.SetN(&frame, MMFR_TA, 0b11, ta)
	bits.SetN(&frame, MMFR_DATA, 0xffff, uint32(data))

	return
}
