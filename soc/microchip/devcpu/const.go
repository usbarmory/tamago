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
	CFG_MODE            = 2
	CFG_STATUS_WORD_POS = 1
	CFG_BYTE_SWAP       = 0
	IFH_LEN             = 36
)

// Frame extraction registers
const (
	XTR              = 0x00
	XTR_GRP_CFG      = XTR + 0x00
	XTR_RD           = XTR + 0x08
	XTR_FRM_PRUNING  = XTR + 0x10
	XTR_DATA_PRESENT = XTR + 0x1c
)

// Special frame error values (little endian)
const (
	RD_EOF_UNUSED_0  = 0x00000080
	RD_EOF_UNUSED_1  = 0x01000080
	RD_EOF_UNUSED_2  = 0x02000080
	RD_EOF_UNUSED_3  = 0x03000080
	RD_EOF_TRUNCATED = 0x04000080
	RD_EOF_ABORTED   = 0x05000080
	RD_ESCAPE        = 0x06000080
	RD_NOT_READY     = 0x07000080
)

// Frame injection registers
const (
	INJ         = 0x24
	INJ_GRP_CFG = INJ + 0x00
	INJ_WR      = INJ + 0x08

	INJ_CTRL       = INJ + 0x10
	CTRL_GAP_SIZE  = 21
	CTRL_ABORT     = 20
	CTRL_EOF       = 19
	CTRL_SOF       = 18
	CTRL_VLD_BYTES = 16

	INJ_STATUS                 = INJ + 0x18
	INJ_STATUS_WMARK_REACHED   = 4
	INJ_STATUS_FIFO_RDY        = 2
	INJ_STATUS_INJ_IN_PROGRESS = 0
)
