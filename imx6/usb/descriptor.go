// USB descriptor support
// https://github.com/f-secure-foundry/tamago
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
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"unicode/utf16"
)

const (
	DEVICE_LENGTH           = 18
	CONFIGURATION_LENGTH    = 9
	INTERFACE_LENGTH        = 9
	ENDPOINT_LENGTH         = 7
	DEVICE_QUALIFIER_LENGTH = 10
)

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
func (d *DeviceDescriptor) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, d)
	return buf.Bytes()
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
func (d *ConfigurationDescriptor) Bytes() []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, d.Length)
	binary.Write(buf, binary.LittleEndian, d.DescriptorType)
	binary.Write(buf, binary.LittleEndian, d.TotalLength)
	binary.Write(buf, binary.LittleEndian, d.NumInterfaces)
	binary.Write(buf, binary.LittleEndian, d.ConfigurationValue)
	binary.Write(buf, binary.LittleEndian, d.Configuration)
	binary.Write(buf, binary.LittleEndian, d.Attributes)
	binary.Write(buf, binary.LittleEndian, d.MaxPower)

	return buf.Bytes()
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

	Endpoints        []*EndpointDescriptor
	ClassDescriptors [][]byte
}

// SetDefaults initializes default values for the USB interface descriptor.
func (d *InterfaceDescriptor) SetDefaults() {
	d.Length = INTERFACE_LENGTH
	d.DescriptorType = INTERFACE
	d.NumEndpoints = 1
}

// Bytes converts the descriptor structure to byte array format.
func (d *InterfaceDescriptor) Bytes() []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, d.Length)
	binary.Write(buf, binary.LittleEndian, d.DescriptorType)
	binary.Write(buf, binary.LittleEndian, d.InterfaceNumber)
	binary.Write(buf, binary.LittleEndian, d.AlternateSetting)
	binary.Write(buf, binary.LittleEndian, d.NumEndpoints)
	binary.Write(buf, binary.LittleEndian, d.InterfaceClass)
	binary.Write(buf, binary.LittleEndian, d.InterfaceSubClass)
	binary.Write(buf, binary.LittleEndian, d.InterfaceProtocol)
	binary.Write(buf, binary.LittleEndian, d.Interface)

	// add class descriptors
	for _, classDesc := range d.ClassDescriptors {
		buf.Write(classDesc)
	}

	return buf.Bytes()
}

// EndpointFunction represents the function to process either IN or OUT
// transfer, depending on the endpoint configuration.
//
// On OUT transfers the function is expected to receive data from the host in
// the `out` byte array, the `in` return value is ignored.
//
// On IN transfers the function is expected to return a data slice for
// transmission to the host, such data is used to fill the DMA buffer in
// advance, to respond to IN requests. The function is invoked by the
// EndpointHandler to fill the buffer as needed.
type EndpointFunction func(out []byte, lastErr error) (in []byte, err error)

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
func (d *EndpointDescriptor) Bytes() []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, d.Length)
	binary.Write(buf, binary.LittleEndian, d.DescriptorType)
	binary.Write(buf, binary.LittleEndian, d.EndpointAddress)
	binary.Write(buf, binary.LittleEndian, d.Attributes)
	binary.Write(buf, binary.LittleEndian, d.MaxPacketSize)
	binary.Write(buf, binary.LittleEndian, d.Interval)

	return buf.Bytes()
}

// StringDescriptor implements
// p273, 9.6.7 String, USB Specification Revision 2.0.
type StringDescriptor struct {
	Length         uint8
	DescriptorType uint8
}

// SetDefaults initializes default values for the USB string descriptor.
func (d *StringDescriptor) SetDefaults() {
	d.Length = 2
	d.DescriptorType = STRING
}

// Bytes converts the descriptor structure to byte array format.
func (d *StringDescriptor) Bytes() []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, d.Length)
	binary.Write(buf, binary.LittleEndian, d.DescriptorType)

	return buf.Bytes()
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
func (d *DeviceQualifierDescriptor) Bytes() []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, d.Length)
	binary.Write(buf, binary.LittleEndian, d.DescriptorType)
	binary.Write(buf, binary.LittleEndian, d.bcdUSB)
	binary.Write(buf, binary.LittleEndian, d.DeviceClass)
	binary.Write(buf, binary.LittleEndian, d.DeviceSubClass)
	binary.Write(buf, binary.LittleEndian, d.DeviceProtocol)
	binary.Write(buf, binary.LittleEndian, d.MaxPacketSize)
	binary.Write(buf, binary.LittleEndian, d.NumConfigurations)
	binary.Write(buf, binary.LittleEndian, d.Reserved)

	return buf.Bytes()
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
