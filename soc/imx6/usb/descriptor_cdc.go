// USB descriptor support
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

// CDC descriptor constants
const (
	// p44, Table 24: Type Values for the bDescriptorType Field,
	// USB Class Definitions for Communication Devices 1.1
	CS_INTERFACE = 0x24

	HEADER_LENGTH              = 5
	UNION_LENGTH               = 5
	ETHERNET_NETWORKING_LENGTH = 13

	// p64, Table 46: Class-Specific Request Codes,
	// USB Class Definitions for Communication Devices 1.1
	SET_ETHERNET_PACKET_FILTER = 0x43

	HEADER              = 0
	UNION               = 6
	ETHERNET_NETWORKING = 15

	// Maximum Segment Size
	MSS = 1500 + 14
)

// CDCHeaderDescriptor implements
// p45, Table 26: Class-Specific Descriptor Header Format, USB Class
// Definitions for Communication Devices 1.1.
type CDCHeaderDescriptor struct {
	Length            uint8
	DescriptorType    uint8
	DescriptorSubType uint8
	bcdCDC            uint16
}

// SetDefaults initializes default values for the USB CDC Header Functional
// Descriptor.
func (d *CDCHeaderDescriptor) SetDefaults() {
	d.Length = HEADER_LENGTH
	d.DescriptorType = CS_INTERFACE
	d.DescriptorSubType = HEADER
	// CDC 1.10
	d.bcdCDC = 0x0110
}

// Bytes converts the descriptor structure to byte array format.
func (d *CDCHeaderDescriptor) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, d)
	return buf.Bytes()
}

// CDCUnionDescriptor implements
// p51, Table 33: Union Interface Functional Descriptor, USB Class Definitions
// for Communication Devices 1.1.
type CDCUnionDescriptor struct {
	Length            uint8
	DescriptorType    uint8
	DescriptorSubType uint8
	MasterInterface   uint8
	SlaveInterface0   uint8
}

// SetDefaults initializes default values for the USB CDC Union Functional
// Descriptor.
func (d *CDCUnionDescriptor) SetDefaults() {
	d.Length = UNION_LENGTH
	d.DescriptorType = CS_INTERFACE
	d.DescriptorSubType = UNION
}

// Bytes converts the descriptor structure to byte array format.
func (d *CDCUnionDescriptor) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, d)
	return buf.Bytes()
}

// CDCEthernetDescriptor implements
// p56, Table 41: Ethernet Networking Functional Descriptor, USB Class
// Definitions for Communication Devices 1.1.
type CDCEthernetDescriptor struct {
	Length             uint8
	DescriptorType     uint8
	DescriptorSubType  uint8
	MacAddress         uint8
	EthernetStatistics uint32
	MaxSegmentSize     uint16
	NumberMCFilters    uint16
	NumberPowerFilters uint8
}

// SetDefaults initializes default values for the USB CDC Ethernet Networking
// Functional Descriptor.
func (d *CDCEthernetDescriptor) SetDefaults() {
	d.Length = ETHERNET_NETWORKING_LENGTH
	d.DescriptorType = CS_INTERFACE
	d.DescriptorSubType = ETHERNET_NETWORKING
	d.MaxSegmentSize = MSS
}

// Bytes converts the descriptor structure to byte array format.
func (d *CDCEthernetDescriptor) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, d)
	return buf.Bytes()
}
