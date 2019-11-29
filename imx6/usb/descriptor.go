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
	"unicode/utf16"
	"unsafe"
)

// p276, Table 9-2. Format of Setup Data, USB Specification Revision 2.0
type SetupData struct {
	bRequestType uint8
	bRequest     uint8
	wValue       uint16
	wIndex       uint16
	wLength      uint16
}

// The endianness values written in memory by the hardware does not match the
// expected one by Go, so we have to swap multi byte values.
func (s *SetupData) swap() {
	b := make([]byte, 2)

	binary.BigEndian.PutUint16(b, s.wValue)
	s.wValue = binary.LittleEndian.Uint16(b)

	binary.BigEndian.PutUint16(b, s.wIndex)
	s.wIndex = binary.LittleEndian.Uint16(b)
}

// p290, Table 9-8. Standard Device Descriptor, USB Specification Revision 2.0
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

// Set default values for USB device descriptor.
func (d *DeviceDescriptor) SetDefaults() {
	d.Length = 18
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

// Convert descriptor structure to byte array format.
func (d *DeviceDescriptor) Bytes() []byte {
	var b [18]byte
	buf := b[:]
	p := uintptr(unsafe.Pointer(&buf[0]))

	desc := (*DeviceDescriptor)(unsafe.Pointer(p))
	*desc = *d

	return buf
}

// p293, Table 9-10. Standard Configuration Descriptor, USB Specification Revision 2.0
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

// Set default values for USB configuration descriptor.
func (d *ConfigurationDescriptor) SetDefaults() {
	d.Length = 9
	d.DescriptorType = CONFIGURATION
	d.NumInterfaces = 1
	d.ConfigurationValue = 1
	d.Attributes = 0xc0
	d.MaxPower = 250
}

// Convert descriptor structure to byte array format.
func (d *ConfigurationDescriptor) Bytes() (buf []byte) {
	buf = append(buf, make([]byte, unsafe.Sizeof(ConfigurationDescriptor{}))...)
	p := uintptr(unsafe.Pointer(&buf[0]))

	desc := (*ConfigurationDescriptor)(unsafe.Pointer(p))
	*desc = *d

	// skip d.Interfaces when returning the buffer
	return buf[0:9]
}

// p296, Table 9-12. Standard Interface Descriptor, USB Specification Revision 2.0
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

// Set default values for USB interface descriptor.
func (d *InterfaceDescriptor) SetDefaults() {
	d.Length = 9
	d.DescriptorType = INTERFACE
	d.NumEndpoints = 1
}

// Convert descriptor structure to byte array format.
func (d *InterfaceDescriptor) Bytes() (buf []byte) {
	buf = append(buf, make([]byte, unsafe.Sizeof(InterfaceDescriptor{}))...)
	p := uintptr(unsafe.Pointer(&buf[0]))

	desc := (*InterfaceDescriptor)(unsafe.Pointer(p))
	*desc = *d

	// skip d.Endpoints when returning the buffer
	return buf[0:9]
}

// p297, Table 9-13. Standard Endpoint Descriptor, USB Specification Revision 2.0
type EndpointDescriptor struct {
	Length          uint8
	DescriptorType  uint8
	EndpointAddress uint8
	Attributes      uint8
	MaxPacketSize   uint16
	Interval        uint8
}

// Set default values for USB endpoint descriptor.
func (d *EndpointDescriptor) SetDefaults() {
	d.Length = 7
	d.DescriptorType = ENDPOINT
	// EP1 IN
	d.EndpointAddress = 0x81
}

// Convert descriptor structure to byte array format.
func (d *EndpointDescriptor) Bytes() (buf []byte) {
	buf = append(buf, make([]byte, unsafe.Sizeof(EndpointDescriptor{}))...)
	p := uintptr(unsafe.Pointer(&buf[0]))

	desc := (*EndpointDescriptor)(unsafe.Pointer(p))
	*desc = *d

	return buf[0:7]
}

// p273, 9.6.7 String, USB Specification Revision 2.0
type StringDescriptor struct {
	Length         uint8
	DescriptorType uint8
}

// Convert descriptor structure to byte array format.
func (d *StringDescriptor) Bytes() (buf []byte) {
	buf = append(buf, make([]byte, unsafe.Sizeof(StringDescriptor{}))...)
	p := uintptr(unsafe.Pointer(&buf[0]))

	desc := (*StringDescriptor)(unsafe.Pointer(p))
	*desc = *d

	return buf[0:2]
}

// Set default values for USB endpoint descriptor.
func (d *StringDescriptor) SetDefaults() {
	d.Length = 2
	d.DescriptorType = STRING
}

// p292, 9.6.2 Device_Qualifier, USB Specification Revision 2.0
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

// Set default values for USB device qualifier descriptor.
func (d *DeviceQualifierDescriptor) SetDefaults() {
	d.Length = 10
	d.DescriptorType = DEVICE_QUALIFIER
	// USB 2.0
	d.bcdUSB = 0x0200
	// maximum packet size for EP0
	d.MaxPacketSize = 64
	d.NumConfigurations = 1
}

// Convert descriptor structure to byte array format.
func (d *DeviceQualifierDescriptor) Bytes() (buf []byte) {
	buf = append(buf, make([]byte, unsafe.Sizeof(DeviceQualifierDescriptor{}))...)
	p := uintptr(unsafe.Pointer(&buf[0]))

	desc := (*DeviceQualifierDescriptor)(unsafe.Pointer(p))
	*desc = *d

	return buf[0:10]
}

// USB device descriptors
type Device struct {
	Descriptor     *DeviceDescriptor
	Qualifier      *DeviceQualifierDescriptor
	Configurations []*ConfigurationDescriptor
	Strings        [][]byte

	// Host requested settings
	ConfigurationValue uint8
	AlternateSetting   uint8
}

// Add a string descriptor, the argument can be an array of integers to create
// String Descriptor Zero or a string to create a UNICODE String Descriptor.
//
// The String Descriptor Zero should always be the first invocation as it
// belongs to index 0, otherwise an error is returned.
//
// The returned index can be used to fill string descriptor index value in
// configuration descriptors.
func (d *Device) AddString(s interface{}) (index uint8, err error) {
	var descBuf []byte
	var dataBuf []byte

	desc := &StringDescriptor{}
	desc.SetDefaults()

	switch s.(type) {
	// p273, Table 9-15. String Descriptor Zero, Specifying Languages Supported by the Device, USB Specification Revision 2.0
	case []uint16:
		if len(d.Strings) != 0 {
			err = fmt.Errorf("string descriptor zero must be added first")
			return
		}

		langs := s.([]uint16)

		if len(langs) > 1 {
			// TODO
			err = fmt.Errorf("only a single language is currently supported")
			return
		}

		for i := 0; i < len(langs); i++ {
			b := make([]byte, 2)
			binary.BigEndian.PutUint16(b, langs[i])
			dataBuf = append(dataBuf, b...)
		}
	// p274, Table 9-16. UNICODE String Descriptor, USB Specification Revision 2.0
	case string:
		if len(d.Strings) == 0 {
			err = fmt.Errorf("string descriptor zero must be added first")
			return
		}

		r := []rune(s.(string))
		u := utf16.Encode([]rune(r))

		for i := 0; i < len(u); i++ {
			dataBuf = append(dataBuf, byte(u[i]&0xff))
			dataBuf = append(dataBuf, byte(u[i]>>8))
		}
	default:
		err = fmt.Errorf("unsupported data type (%T)", s)
		return
	}

	desc.Length += uint8(len(dataBuf))

	if desc.Length > 255 {
		err = fmt.Errorf("string descriptor size (%d) cannot exceed 255", desc.Length)
		return
	}

	descBuf = append(descBuf, desc.Bytes()...)
	descBuf = append(descBuf, dataBuf...)

	d.Strings = append(d.Strings, descBuf)
	index = uint8(len(d.Strings) - 1)

	return
}

// Convert entire configuration hierarchy to a buffer, as expected by Get
// Descriptor for configuration descriptor type.
//
// p281, 9.4.3 Get Descriptor, USB Specification Revision 2.0
func (d *Device) Configuration(wIndex uint16, wLength uint16) (buf []byte, err error) {
	if int(wIndex+1) > len(d.Configurations) {
		err = errors.New("invalid configuration index")
		return
	}

	conf := d.Configurations[int(wIndex)]
	buf = append(buf, conf.Bytes()...)

	if int(wLength) <= len(buf) {
		return
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
