// Microchip CPU port module (DEVCPU)
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package devcpu

// Common fields
const (
	CFG_MODE = 2
	IFH_LEN  = 36
)

// Frame extraction registers
const (
	XTR              = 0x00
	XTR_GRP_CFG      = XTR + 0x00
	XTR_RD           = XTR + 0x08
	XTR_DATA_PRESENT = XTR + 0x1c
)

// Special frame error values (little endian)
const (
	RD_EOF_UNUSED_0  = 0x00000080
	RD_EOF_UNUSED_1  = 0x10000080
	RD_EOF_UNUSED_2  = 0x20000080
	RD_EOF_UNUSED_3  = 0x30000080
	RD_EOF_TRUNCATED = 0x40000080
	RD_EOF_ABORTED   = 0x50000080
	RD_ESCAPE        = 0x60000080
	RD_NOT_READY     = 0x70000080
)

// Frame injection registers
const (
	INJ         = 0x24
	INJ_GRP_CFG = INJ + 0x00
	INJ_WR      = INJ + 0x08

	INJ_CTRL       = INJ + 0x10
	CTRL_GAP_SIZE  = 21
	CTRL_EOF       = 19
	CTRL_SOF       = 18
	CTRL_VLD_BYTES = 16
)
