// USB mass storage descriptor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package usb

import (
	"bytes"
	"encoding/binary"
)

// Mass Storage constants
const (
	// p11, Table 4.5 - Bulk-Only Data Interface Descriptor,
	// USB Mass Storage Class 1.0
	MASS_STORAGE_CLASS           = 0x08
	BULK_ONLY_TRANSPORT_PROTOCOL = 0x50

	// p11, Table 1 — SubClass Codes Mapped to Command Block
	// Specifications, MSCO Revision 1.4
	SCSI_CLASS = 0x06

	CBW_LENGTH        = 31
	CBW_CB_MAX_LENGTH = 16

	CBW_SIGNATURE = 0x43425355
	CSW_SIGNATURE = 0x53425355

	// p15, Table 5.3 - Command Block Status Values,
	// USB Mass Storage Class 1.0
	CSW_STATUS_COMMAND_PASSED = 0x00
	CSW_STATUS_COMMAND_FAILED = 0x01
	CSW_STATUS_PHASE_ERROR    = 0x02

	// p7, 3.1 - 3.2, USB Mass Storage Class 1.0
	BULK_ONLY_MASS_STORAGE_RESET = 0xff
	GET_MAX_LUN                  = 0xfe
)

// CBW implements p13, 5.1 Command Block Wrapper (CBW),
// USB Mass Storage Class 1.0
type CBW struct {
	Signature          uint32
	Tag                uint32
	DataTransferLength uint32
	Flags              uint8
	LUN                uint8
	Length             uint8
	CommandBlock       [16]byte
}

// SetDefaults initializes default values for the CBW descriptor.
func (d *CBW) SetDefaults() {
	d.Signature = CBW_SIGNATURE
}

// Bytes converts the descriptor structure to byte array format.
func (d *CBW) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, d)
	return buf.Bytes()
}

// CSW implements p14, 5.2 Command Status Wrapper (CSW), USB Mass Storage Class 1.0
type CSW struct {
	Signature   uint32
	Tag         uint32
	DataResidue uint32
	Status      uint8
}

// SetDefaults initializes default values for the CSW descriptor.
func (d *CSW) SetDefaults() {
	d.Signature = CSW_SIGNATURE
	d.Status = CSW_STATUS_COMMAND_PASSED
}

// Bytes converts the descriptor structure to byte array format.
func (d *CSW) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, d)
	return buf.Bytes()
}
