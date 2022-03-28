// CCID descriptor support
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

// CCID descriptor constants
const (
	// p17, Table 5.1-1, CCID Rev1.1
	CCID_INTERFACE         = 0x21
	CCID_DESCRIPTOR_LENGTH = 54
)

// CCIDDescriptor implements p17, Table 5.1-1, CCID Rev1.1
type CCIDDescriptor struct {
	Length                uint8
	DescriptorType        uint8
	CCID                  uint16
	MaxSlotIndex          uint8
	VoltageSupport        uint8
	Protocols             uint32
	DefaultClock          uint32
	MaximumClock          uint32
	NumClockSupported     uint8
	DataRate              uint32
	MaxDataRate           uint32
	NumDataRatesSupported uint8
	MaxIFSD               uint32
	SynchProtocols        uint32
	Mechanical            uint32
	Features              uint32
	MaxCCIDMessageLength  uint32
	ClassGetResponse      uint8
	ClassEnvelope         uint8
	LcdLayout             uint16
	PINSupport            uint8
	MaxCCIDBusySlots      uint8
}

// SetDefaults initializes default values for the USB Smart Card Device Class
// Descriptor.
func (d *CCIDDescriptor) SetDefaults() {
	d.Length = CCID_DESCRIPTOR_LENGTH
	d.DescriptorType = CCID_INTERFACE
	d.CCID = 0x0110
	// all voltages
	d.VoltageSupport = 0x7
	// T=1
	d.Protocols = 0x2
	d.DefaultClock = 0x4000
	d.MaximumClock = 0x4000
	d.DataRate = 0x4b000
	d.MaxDataRate = 0x4b000
	d.MaxIFSD = 0xfe
	// Features:
	//   Auto configuration based on ATR
	//   Auto activation on insert
	//   Auto voltage selection
	//   Auto clock change
	//   Auto baud rate change
	//   Auto parameter negotiation made by CCID
	//   Short and extended APDU level exchange
	d.Features = 0x400fe
	d.MaxCCIDMessageLength = DTD_PAGES * DTD_PAGE_SIZE
	// echo
	d.ClassGetResponse = 0xff
	d.ClassEnvelope = 0xff
	d.MaxCCIDBusySlots = 1
}

// Bytes converts the descriptor structure to byte array format.
func (d *CCIDDescriptor) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, d)
	return buf.Bytes()
}
