// USB descriptor support
// https://github.com/inversepath/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build tamago,arm

package usb

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"unicode/utf16"
	"unsafe"
)

const (
	DEVICE_LENGTH           = 18
	CONFIGURATION_LENGTH    = 9
	INTERFACE_LENGTH        = 9
	ENDPOINT_LENGTH         = 7
	DEVICE_QUALIFIER_LENGTH = 10
)

// SetupData implements
// p276, Table 9-2. Format of Setup Data, USB Specification Revision 2.0.
type SetupData struct {
	bRequestType uint8
	bRequest     uint8
	wValue       uint16
	wIndex       uint16
	wLength      uint16
}

// swap adjusts the endianness of values written in memory by the hardware, as
// they do not match the expected one by Go.
func (s *SetupData) swap() {
	b := make([]byte, 2)

	binary.BigEndian.PutUint16(b, s.wValue)
	s.wValue = binary.LittleEndian.Uint16(b)

	binary.BigEndian.PutUint16(b, s.wIndex)
	s.wIndex = binary.LittleEndian.Uint16(b)
}

// DeviceDescriptor implements
// p290, Table 9-8. Standard Device Descriptor, USB Specification Revision 2.0.
type DeviceDescriptor struct {
	Length            uint8
	DescriptorType    uint8
	bcdUSB            uint16
	DeviceClass       uint8
	DeviceSubClass    uint8
	DeviceProtocol    uint8
	MaxPacketSize     uint8
	VendorId          uint16
	ProductId         uint16
	Device            uint16
	Manufacturer      uint8
	Product           uint8
	SerialNumber      uint8
	NumConfigurations uint8
}

// SetDefaults initializes default values for the USB device descriptor.
func (d *DeviceDescriptor) SetDefaults() {
	d.Length = DEVICE_LENGTH
	d.DescriptorType = DEVICE
	// USB 2.0
	d.bcdUSB = 0x0200
	// maximum packet size for EP0
	d.MaxPacketSize = 64
	// http://pid.codes/1209/2702/
	d.VendorId = 0x1209
	d.ProductId = 0x2702
	d.NumConfigurations = 1
}

// Bytes converts the descriptor structure to byte array format.
func (d *DeviceDescriptor) Bytes() (buf []byte) {
	size := unsafe.Sizeof(*d)
	buf = make([]byte, size, size)
	p := uintptr(unsafe.Pointer(&buf[0]))

	desc := (*DeviceDescriptor)(unsafe.Pointer(p))
	*desc = *d

	return buf[0:DEVICE_LENGTH]
}

// ConfigurationDescriptor implements
// p293, Table 9-10. Standard Configuration Descriptor, USB Specification Revision 2.0.
type ConfigurationDescriptor struct {
	Length             uint8
	DescriptorType     uint8
	TotalLength        uint16
	NumInterfaces      uint8
	ConfigurationValue uint8
	Configuration      uint8
	Attributes         uint8
	MaxPower           uint8

	Interfaces []*InterfaceDescriptor
}

// SetDefaults initializes default values for the USB configuration descriptor.
func (d *ConfigurationDescriptor) SetDefaults() {
	d.Length = CONFIGURATION_LENGTH
	d.DescriptorType = CONFIGURATION
	d.NumInterfaces = 1
	d.ConfigurationValue = 1
	d.Attributes = 0xc0
	d.MaxPower = 250
}

// Bytes converts the descriptor structure to byte array format.
func (d *ConfigurationDescriptor) Bytes() (buf []byte) {
	size := unsafe.Sizeof(*d)
	buf = make([]byte, size, size)
	p := uintptr(unsafe.Pointer(&buf[0]))

	desc := (*ConfigurationDescriptor)(unsafe.Pointer(p))
	*desc = *d

	// skip d.Interfaces when returning the buffer
	return buf[0:CONFIGURATION_LENGTH]
}

// InterfaceDescriptor implements
// p296, Table 9-12. Standard Interface Descriptor, USB Specification Revision 2.0.
type InterfaceDescriptor struct {
	Length            uint8
	DescriptorType    uint8
	InterfaceNumber   uint8
	AlternateSetting  uint8
	NumEndpoints      uint8
	InterfaceClass    uint8
	InterfaceSubClass uint8
	InterfaceProtocol uint8
	Interface         uint8

	Endpoints []*EndpointDescriptor
}

// SetDefaults initializes default values for the USB interface descriptor.
func (d *InterfaceDescriptor) SetDefaults() {
	d.Length = INTERFACE_LENGTH
	d.DescriptorType = INTERFACE
	d.NumEndpoints = 1
}

// Bytes converts the descriptor structure to byte array format.
func (d *InterfaceDescriptor) Bytes() (buf []byte) {
	size := unsafe.Sizeof(*d)
	buf = make([]byte, size, size)
	p := uintptr(unsafe.Pointer(&buf[0]))

	desc := (*InterfaceDescriptor)(unsafe.Pointer(p))
	*desc = *d

	// skip d.Endpoints when returning the buffer
	return buf[0:INTERFACE_LENGTH]
}

type EndpointFunction func(max uint16) (data []byte, err error)

// EndpointDescriptor implements
// p297, Table 9-13. Standard Endpoint Descriptor, USB Specification Revision 2.0.
type EndpointDescriptor struct {
	Length          uint8
	DescriptorType  uint8
	EndpointAddress uint8
	Attributes      uint8
	MaxPacketSize   uint16
	Interval        uint8

	Function EndpointFunction

	enabled bool
	sync.Mutex
}

// SetDefaults initializes default values for the USB endpoint descriptor.
func (d *EndpointDescriptor) SetDefaults() {
	d.Length = ENDPOINT_LENGTH
	d.DescriptorType = ENDPOINT
	// EP1 IN
	d.EndpointAddress = 0x81
}

// Number returns the endpoint number.
func (d *EndpointDescriptor) Number() int {
	return int(d.EndpointAddress & 0b1111)
}

// Direction returns the endpoint direction.
func (d *EndpointDescriptor) Direction() int {
	return int(d.EndpointAddress&0b10000000) / 0b10000000
}

// TransferType returns the endpoint transfer type.
func (d *EndpointDescriptor) TransferType() int {
	return int(d.Attributes & 0b11)
}

// Bytes converts the descriptor structure to byte array format.
func (d *EndpointDescriptor) Bytes() (buf []byte) {
	size := unsafe.Sizeof(*d)
	buf = make([]byte, size, size)
	p := uintptr(unsafe.Pointer(&buf[0]))

	desc := (*EndpointDescriptor)(unsafe.Pointer(p))
	*desc = *d

	return buf[0:ENDPOINT_LENGTH]
}

// StringDescriptor implements
// p273, 9.6.7 String, USB Specification Revision 2.0.
type StringDescriptor struct {
	Length         uint8
	DescriptorType uint8
}

// SetDefaults initializes default values for the USB string descriptor.
func (d *StringDescriptor) SetDefaults() {
	d.Length = uint8(unsafe.Sizeof(*d))
	d.DescriptorType = STRING
}

// Bytes converts the descriptor structure to byte array format.
func (d *StringDescriptor) Bytes() (buf []byte) {
	size := unsafe.Sizeof(*d)
	buf = make([]byte, size, size)
	p := uintptr(unsafe.Pointer(&buf[0]))

	desc := (*StringDescriptor)(unsafe.Pointer(p))
	*desc = *d

	return buf
}

// DeviceQualifierDescriptor implements
// p292, 9.6.2 Device_Qualifier, USB Specification Revision 2.0.
type DeviceQualifierDescriptor struct {
	Length            uint8
	DescriptorType    uint8
	bcdUSB            uint16
	DeviceClass       uint8
	DeviceSubClass    uint8
	DeviceProtocol    uint8
	MaxPacketSize     uint8
	NumConfigurations uint8
	Reserved          uint8
}

// SetDefaults initializes default values for the USB device qualifier
// descriptor.
func (d *DeviceQualifierDescriptor) SetDefaults() {
	d.Length = DEVICE_QUALIFIER_LENGTH
	d.DescriptorType = DEVICE_QUALIFIER
	// USB 2.0
	d.bcdUSB = 0x0200
	// maximum packet size for EP0
	d.MaxPacketSize = 64
	d.NumConfigurations = 1
}

// Bytes converts the descriptor structure to byte array format.
func (d *DeviceQualifierDescriptor) Bytes() (buf []byte) {
	size := unsafe.Sizeof(*d)
	buf = make([]byte, size, size)
	p := uintptr(unsafe.Pointer(&buf[0]))

	desc := (*DeviceQualifierDescriptor)(unsafe.Pointer(p))
	*desc = *d

	return buf[0:DEVICE_QUALIFIER_LENGTH]
}

// Device is a collection of USB device descriptors and host driven settings
// to represent a USB device.
type Device struct {
	Descriptor     *DeviceDescriptor
	Qualifier      *DeviceQualifierDescriptor
	Configurations []*ConfigurationDescriptor
	Strings        [][]byte

	// Host requested settings
	ConfigurationValue uint8
	AlternateSetting   uint8
}

func (d *Device) setStringDescriptor(s []byte, zero bool) (uint8, error) {
	var buf []byte

	desc := &StringDescriptor{}
	desc.SetDefaults()
	desc.Length += uint8(len(s))

	if desc.Length > 255 {
		return 0, fmt.Errorf("string descriptor size (%d) cannot exceed 255", desc.Length)
	}

	buf = append(buf, desc.Bytes()...)
	buf = append(buf, s...)

	if zero && len(d.Strings) >= 1 {
		d.Strings[0] = buf
	} else {
		d.Strings = append(d.Strings, buf)
	}

	return uint8(len(d.Strings) - 1), nil
}

// SetLanguageCodes configures String Descriptor Zero language codes
// (p273, Table 9-15. String Descriptor Zero, Specifying Languages Supported by the Device, USB Specification Revision 2.0).
func (d *Device) SetLanguageCodes(codes []uint16) (err error) {
	var buf []byte

	if len(codes) > 1 {
		// TODO
		return fmt.Errorf("only a single language is currently supported")
	}

	for i := 0; i < len(codes); i++ {
		b := make([]byte, 2)
		binary.BigEndian.PutUint16(b, codes[i])
		buf = append(buf, b...)
	}

	_, err = d.setStringDescriptor(buf, true)

	return
}

// AddStrings adds a string descriptor to a USB device. The returned index can
// be used to fill string descriptor index value in configuration descriptors
// (p274, Table 9-16. UNICODE String Descriptor, USB Specification Revision 2.0).
func (d *Device) AddString(s string) (uint8, error) {
	var buf []byte

	desc := &StringDescriptor{}
	desc.SetDefaults()

	r := []rune(s)
	u := utf16.Encode([]rune(r))

	for i := 0; i < len(u); i++ {
		buf = append(buf, byte(u[i]&0xff))
		buf = append(buf, byte(u[i]>>8))
	}

	return d.setStringDescriptor(buf, false)
}

// Configuration converts the device configuration hierarchy to a buffer, as expected by Get
// Descriptor for configuration descriptor type
// (p281, 9.4.3 Get Descriptor, USB Specification Revision 2.0).
func (d *Device) Configuration(wIndex uint16, wLength uint16) (buf []byte, err error) {
	if int(wIndex+1) > len(d.Configurations) {
		err = errors.New("invalid configuration index")
		return
	}

	conf := d.Configurations[int(wIndex)]
	buf = append(buf, conf.Bytes()...)

	if int(wLength) <= len(buf) {
		return buf, err
	}

	for i := 0; i < len(conf.Interfaces); i++ {
		iface := conf.Interfaces[i]
		buf = append(buf, iface.Bytes()...)

		for i := 0; i < len(iface.Endpoints); i++ {
			ep := iface.Endpoints[i]
			buf = append(buf, ep.Bytes()...)
		}
	}

	if int(wLength) > len(buf) {
		// device may return less than what is requested
		return buf, err
	}

	return buf[0:wLength], err
}
