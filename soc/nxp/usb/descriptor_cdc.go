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
	// p39, Table 14: Communication Device Class Code
	// USB Class Definitions for Communication Devices 1.1
	COMMUNICATION_DEVICE_CLASS = 0x02

	// p39, Table 15: Communication Interface Class Code
	// USB Class Definitions for Communication Devices 1.1
	COMMUNICATION_INTERFACE_CLASS = 0x02

	// p40, Table 17: Data Interface Class Code
	// USB Class Definitions for Communication Devices 1.1
	DATA_INTERFACE_CLASS = 0x0a

	// p44, Table 24: Type Values for the bDescriptorType Field,
	// USB Class Definitions for Communication Devices 1.1
	CS_INTERFACE = 0x24

	// p64, Table 46: Class-Specific Request Codes,
	// USB Class Definitions for Communication Devices 1.1
	SET_ETHERNET_PACKET_FILTER = 0x43

	// Maximum Segment Size
	MSS = 1500 + 14
)

// p39, Table 16: Communication Interface Class SubClass Codes,
// USB Class Definitions for Communication Devices 1.1
const (
	ACM_SUBCLASS = 0x02
	ETH_SUBCLASS = 0x06
)

// p40, Table 17: Communication Interface Class Control Protocol Codes,
// USB Class Definitions for Communication Devices 1.1
const (
	AT_COMMAND_PROTOCOL = 0x01
)

// p44, Table 25: bDescriptor SubType in Functional Descriptors,
// USB Class Definitions for Communication Devices 1.1
const (
	HEADER                      = 0x00
	CALL_MANAGEMENT             = 0x01
	ABSTRACT_CONTROL_MANAGEMENT = 0x02
	UNION                       = 0x06
	ETHERNET_NETWORKING         = 0x0f
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
	d.Length = 5
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

// CDCCallManagementDescriptor implements
// p45, Table 27: Call Management Functional Descriptor, USB Class Definitions
// for Communication Devices 1.1.
type CDCCallManagementDescriptor struct {
	Length            uint8
	DescriptorType    uint8
	DescriptorSubType uint8
	Capabilities      uint8
	DataInterface     uint8
}

// SetDefaults initializes default values for the USB CDC Call Management
// Descriptor.
func (d *CDCCallManagementDescriptor) SetDefaults() {
	d.Length = 5
	d.DescriptorType = CS_INTERFACE
	d.DescriptorSubType = CALL_MANAGEMENT
}

// Bytes converts the descriptor structure to byte array format.
func (d *CDCCallManagementDescriptor) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, d)
	return buf.Bytes()
}

// CDCAbstractCallManagementDescriptor implements
// p46, Table 28: Abstract Control Management Functional Descriptor, USB Class
// Definitions for Communication Devices 1.1.
type CDCAbstractControlManagementDescriptor struct {
	Length            uint8
	DescriptorType    uint8
	DescriptorSubType uint8
	Capabilities      uint8
}

// SetDefaults initializes default values for the USB CDC Abstract Control
// Management Descriptor.
func (d *CDCAbstractControlManagementDescriptor) SetDefaults() {
	d.Length = 4
	d.DescriptorType = CS_INTERFACE
	d.DescriptorSubType = ABSTRACT_CONTROL_MANAGEMENT
}

// Bytes converts the descriptor structure to byte array format.
func (d *CDCAbstractControlManagementDescriptor) Bytes() []byte {
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
	d.Length = 5
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
	d.Length = 13
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
