// https://github.com/f-secure-foundry/tamago
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
	"fmt"

	"github.com/f-secure-foundry/tamago/imx6/usb"
)

func configureZeroDevice(device *usb.Device) {
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
}

func configureSourceSink(device *usb.Device) {
	// source and sink configuration
	conf := &usb.ConfigurationDescriptor{}
	conf.SetDefaults()
	conf.TotalLength = 32
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

	// source EP1 IN endpoint (bulk)
	ep1IN := &usb.EndpointDescriptor{}
	ep1IN.SetDefaults()
	ep1IN.EndpointAddress = 0x81
	ep1IN.Attributes = 2
	ep1IN.MaxPacketSize = 512
	ep1IN.Function = source

	iface.Endpoints = append(iface.Endpoints, ep1IN)

	// sink EP1 OUT endpoint (bulk)
	ep1OUT := &usb.EndpointDescriptor{}
	ep1OUT.SetDefaults()
	ep1OUT.EndpointAddress = 0x01
	ep1OUT.Attributes = 2
	ep1OUT.MaxPacketSize = 512
	ep1OUT.Function = sink

	iface.Endpoints = append(iface.Endpoints, ep1OUT)
}

// Linux tools/usb/testusb.c does not seem to test loopback functionality at
// all, for now we leave endpoint functions undefined.
func configureLoopback(device *usb.Device) {
	// loopback configuration
	conf := &usb.ConfigurationDescriptor{}
	conf.SetDefaults()
	conf.TotalLength = 0x0020
	conf.NumInterfaces = 1
	conf.ConfigurationValue = 2

	iConfiguration, _ := device.AddString(`loop input to output`)
	conf.Configuration = iConfiguration

	device.Configurations = append(device.Configurations, conf)

	// loopback interface
	iface := &usb.InterfaceDescriptor{}
	iface.SetDefaults()
	iface.NumEndpoints = 2
	iface.InterfaceClass = 0xff

	iInterface, _ := device.AddString(`loop input to output`)
	iface.Interface = iInterface

	conf.Interfaces = append(conf.Interfaces, iface)

	// loopback EP1 IN endpoint (bulk)
	ep1IN := &usb.EndpointDescriptor{}
	ep1IN.SetDefaults()
	ep1IN.EndpointAddress = 0x81
	ep1IN.Attributes = 2
	ep1IN.MaxPacketSize = 512

	iface.Endpoints = append(iface.Endpoints, ep1IN)

	// loopback EP1 OUT endpoint (bulk)
	ep1OUT := &usb.EndpointDescriptor{}
	ep1OUT.SetDefaults()
	ep1OUT.EndpointAddress = 0x01
	ep1OUT.Attributes = 2
	ep1OUT.MaxPacketSize = 512

	iface.Endpoints = append(iface.Endpoints, ep1OUT)
}

// source implements the IN endpoint data source, to be used `modprobe usbtest
// pattern=1 mod_pattern=1`.
func source(_ []byte, lastErr error) (in []byte, err error) {
	in = make([]byte, 512*10)

	for i := 0; i < len(in); i++ {
		in[i] = byte((i % 512) % 63)
	}

	return
}

// sink implemente the OUT endpoint data sink, to be used `modprobe usbtest
// pattern=1 mod_pattern=1`.
func sink(out []byte, lastErr error) (_ []byte, err error) {
	// skip zero length packets
	if len(out) == 0 {
		return
	}

	for i := 0; i < len(out); i++ {
		if out[i] != byte((i%512)%63) {
			return nil, fmt.Errorf("imx6_usb: EP1.0 function error, buffer mismatch (out[%d] == %x)", i, out[i])
		}
	}

	return
}

// StartUSBGadgetZero starts an emulated Linux Gadget Zero device
// (bulk/interrupt endpoints only).
//
// https://github.com/torvalds/linux/blob/master/drivers/usb/gadget/legacy/zero.c
//
// To be tested on host side with `modprobe usbtest pattern=1 mod_pattern=1`.
//
// Example of tests (using Linux tools/usb/testusb.c) expected to pass:
//
// test 0,    0.000007 secs
// test 1,    0.000475 secs
// test 2,    0.000079 secs
// test 3,    0.001594 secs
// test 4,    0.000129 secs
// test 5,    0.011356 secs
// test 6,    0.007847 secs
// test 7,    0.014690 secs
// test 8,    0.007832 secs
// test 10,   0.020543 secs
// test 11,   0.025884 secs
// test 12,   0.029996 secs
// test 17,   0.000058 secs
// test 18,   0.000079 secs
// test 19,   0.001588 secs
// test 20,   0.000092 secs
// test 24,   0.019632 secs
func StartUSBGadgetZero() {
	device := &usb.Device{}

	configureZeroDevice(device)
	configureSourceSink(device)
	configureLoopback(device)

	usb.USB1.Init()
	usb.USB1.DeviceMode()
	usb.USB1.Reset()

	// never returns
	usb.USB1.Start(device)
}
