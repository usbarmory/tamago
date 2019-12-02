// https://github.com/inversepath/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build tamago,arm

package main

import (
	"time"

	"github.com/inversepath/tamago/imx6/usb"
)

// Test function which emulates Linux Gadget Zero descriptors (no actual
// function implemented yet).
func TestUSB() {
	usb.USB1.Init()
	usb.USB1.DeviceMode()

	device := &usb.Device{}

	// https://github.com/torvalds/linux/blob/master/drivers/usb/gadget/legacy/zero.c

	// Supported Language Code Zero: English
	device.SetLanguageCodes([]uint16{0x0409})

	// device descriptor
	device.Descriptor = &usb.DeviceDescriptor{}
	device.Descriptor.SetDefaults()
	device.Descriptor.DeviceClass = 0xff
	device.Descriptor.VendorId = 0x0525
	device.Descriptor.ProductId = 0xa4a0
	device.Descriptor.Device = 0x0001
	device.Descriptor.NumConfigurations = 2

	iManufacturer, _ := device.AddString(`TamaGo`)
	device.Descriptor.Manufacturer = iManufacturer

	iProduct, _ := device.AddString(`Gadget Zero`)
	device.Descriptor.Product = iProduct

	iSerial, _ := device.AddString(`0.1`)
	device.Descriptor.SerialNumber = iSerial

	// device qualifier
	device.Qualifier = &usb.DeviceQualifierDescriptor{}
	device.Qualifier.SetDefaults()
	device.Qualifier.DeviceClass = 0xff
	device.Qualifier.NumConfigurations = 2

	// source and sink configuration
	conf := &usb.ConfigurationDescriptor{}
	conf.SetDefaults()
	conf.TotalLength = 0x0045
	conf.NumInterfaces = 1
	conf.ConfigurationValue = 3

	iConfiguration, _ := device.AddString(`source and sink data`)
	conf.Configuration = iConfiguration

	device.Configurations = append(device.Configurations, conf)

	// source and sink interface
	iface := &usb.InterfaceDescriptor{}
	iface.SetDefaults()
	iface.NumEndpoints = 2
	iface.InterfaceClass = 0xff
	iface.Interface = 0

	conf.Interfaces = append(conf.Interfaces, iface)

	// source and sink EP1 IN endpoint (bulk)
	ep1IN := &usb.EndpointDescriptor{}
	ep1IN.SetDefaults()
	ep1IN.EndpointAddress = 0x81
	ep1IN.Attributes = 2
	ep1IN.MaxPacketSize = 512

	iface.Endpoints = append(iface.Endpoints, ep1IN)

	// source and sink EP1 OUT endpoint (bulk)
	ep1OUT := &usb.EndpointDescriptor{}
	ep1OUT.SetDefaults()
	ep1OUT.EndpointAddress = 0x01
	ep1OUT.Attributes = 2
	ep1OUT.MaxPacketSize = 512

	iface.Endpoints = append(iface.Endpoints, ep1OUT)

	// source and sink alternate interface
	iface = &usb.InterfaceDescriptor{}
	iface.SetDefaults()
	iface.AlternateSetting = 1
	iface.NumEndpoints = 4
	iface.InterfaceClass = 0xff
	iface.Interface = 0

	conf.Interfaces = append(conf.Interfaces, iface)

	// source and sink EP1 IN endpoint (bulk)
	ep1IN = &usb.EndpointDescriptor{}
	ep1IN.SetDefaults()
	ep1IN.EndpointAddress = 0x81
	ep1IN.Attributes = 2
	ep1IN.MaxPacketSize = 512

	iface.Endpoints = append(iface.Endpoints, ep1IN)

	// source and sink EP1 OUT endpoint (bulk)
	ep1OUT = &usb.EndpointDescriptor{}
	ep1OUT.SetDefaults()
	ep1OUT.EndpointAddress = 0x01
	ep1OUT.Attributes = 2
	ep1OUT.MaxPacketSize = 512

	iface.Endpoints = append(iface.Endpoints, ep1OUT)

	// source and sink EP7 IN endpoint (isochronous)
	ep7IN := &usb.EndpointDescriptor{}
	ep7IN.SetDefaults()
	ep7IN.EndpointAddress = 0x87
	ep7IN.Attributes = 1
	ep7IN.MaxPacketSize = 1024
	ep7IN.Interval = 4

	iface.Endpoints = append(iface.Endpoints, ep7IN)

	// source and sink EP2 OUT endpoint (isochronous)
	ep2OUT := &usb.EndpointDescriptor{}
	ep2OUT.SetDefaults()
	ep2OUT.EndpointAddress = 0x02
	ep2OUT.Attributes = 1
	ep2OUT.MaxPacketSize = 1024
	ep2OUT.Interval = 4

	iface.Endpoints = append(iface.Endpoints, ep2OUT)

	// loopback configuration
	conf = &usb.ConfigurationDescriptor{}
	conf.SetDefaults()
	conf.TotalLength = 0x0020
	conf.NumInterfaces = 1
	conf.ConfigurationValue = 2

	iConfiguration, _ = device.AddString(`loop input to output`)
	conf.Configuration = iConfiguration

	device.Configurations = append(device.Configurations, conf)

	// source and sink interface
	iface = &usb.InterfaceDescriptor{}
	iface.SetDefaults()
	iface.NumEndpoints = 2
	iface.InterfaceClass = 0xff

	iInterface, _ := device.AddString(`loop input to output`)
	iface.Interface = iInterface

	conf.Interfaces = append(conf.Interfaces, iface)

	// source and sink EP1 IN endpoint (bulk)
	ep1IN = &usb.EndpointDescriptor{}
	ep1IN.SetDefaults()
	ep1IN.EndpointAddress = 0x81
	ep1IN.Attributes = 2
	ep1IN.MaxPacketSize = 512

	iface.Endpoints = append(iface.Endpoints, ep1IN)

	// source and sink EP1 OUT endpoint (bulk)
	ep1OUT = &usb.EndpointDescriptor{}
	ep1OUT.SetDefaults()
	ep1OUT.EndpointAddress = 0x01
	ep1OUT.Attributes = 2
	ep1OUT.MaxPacketSize = 512

	iface.Endpoints = append(iface.Endpoints, ep1OUT)

	// never returns
	loop := true
	timeout := 5 * time.Second
	usb.USB1.SetupHandler(device, timeout, loop)
}
