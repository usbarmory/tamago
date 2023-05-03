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
	// p16, Table 4.3-1, CCID Rev1.1
	SMARTCARD_DEVICE_CLASS = 0x0b

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
	d.Protocols = 0x3 // support T=0 and T=1

	d.DefaultClock = 4000  // 4 MHz
	d.MaximumClock = 5000  // 5 MHz
	d.DataRate = 9600      // default on power-up
	d.MaxDataRate = 625000 // maximum@5MHz according to ISO7816-3
	// Features:
	//   0x02 Auto configuration based on ATR
	//   0x04 Auto activation on insert
	//   0x08 Auto voltage selection
	//   0x10 Auto clock change
	//   0x20 Auto baud rate change
	//   0x40 Auto parameter negotiation made by CCID
	//   0x40000 Short and extended APDU level exchange
	d.Features = 0x4007E // 0x40 and 0x80 cannot be present at same time
	d.MaxCCIDMessageLength = DTD_PAGES * DTD_PAGE_SIZE
	d.MaxIFSD = d.MaxCCIDMessageLength // max block size = max message size
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
