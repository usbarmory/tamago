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
	"fmt"

	"github.com/inversepath/tamago/imx6/usb"
)

func configureDevice(device *usb.Device) {
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

	// source EP1 IN endpoint (bulk)
	ep1IN := &usb.EndpointDescriptor{}
	ep1IN.SetDefaults()
	ep1IN.EndpointAddress = 0x81
	ep1IN.Attributes = 2
	ep1IN.MaxPacketSize = 512

	// IN data source, to be used `modprobe usbtest pattern=1 mod_pattern=1`
	ep1IN.Function = func(out []byte, lastErr error) (in []byte, err error) {
		in = make([]byte, 512*10)

		for i := 0; i < len(in); i++ {
			in[i] = byte((i % 512) % 63)
		}

		return
	}

	iface.Endpoints = append(iface.Endpoints, ep1IN)

	// sink EP1 OUT endpoint (bulk)
	ep1OUT := &usb.EndpointDescriptor{}
	ep1OUT.SetDefaults()
	ep1OUT.EndpointAddress = 0x01
	ep1OUT.Attributes = 2
	ep1OUT.MaxPacketSize = 512

	// OUT data sink, to be used `modprobe usbtest pattern=1 mod_pattern=1`
	ep1OUT.Function = func(out []byte, lastErr error) (in []byte, err error) {
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

	iface.Endpoints = append(iface.Endpoints, ep1OUT)

	// FIXME: test and implement functions for ep7IN and ep2OUT

	// source and sink alternate interface
	iface = &usb.InterfaceDescriptor{}
	iface.SetDefaults()
	iface.AlternateSetting = 1
	iface.NumEndpoints = 4
	iface.InterfaceClass = 0xff
	iface.Interface = 0

	conf.Interfaces = append(conf.Interfaces, iface)

	// re-use previous descriptors
	iface.Endpoints = append(iface.Endpoints, ep1IN)
	iface.Endpoints = append(iface.Endpoints, ep1OUT)

	// source EP7 IN endpoint (isochronous)
	ep7IN := &usb.EndpointDescriptor{}
	ep7IN.SetDefaults()
	ep7IN.EndpointAddress = 0x87
	ep7IN.Attributes = 1
	ep7IN.MaxPacketSize = 1024
	ep7IN.Interval = 4

	// re-use previous function
	//ep7IN.Function = ep1IN.Function

	iface.Endpoints = append(iface.Endpoints, ep7IN)

	// sink EP2 OUT endpoint (isochronous)
	ep2OUT := &usb.EndpointDescriptor{}
	ep2OUT.SetDefaults()
	ep2OUT.EndpointAddress = 0x02
	ep2OUT.Attributes = 1
	ep2OUT.MaxPacketSize = 1024
	ep2OUT.Interval = 4

	// re-use previous function
	//ep2OUT.Function = ep2OUT.Function

	iface.Endpoints = append(iface.Endpoints, ep2OUT)
}

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

// Test function which emulates Linux Gadget Zero descriptors. To be tested on
// host side with `modprobe usbtest pattern=1 mod_pattern=1`.
func StartUSBGadgetZero() {
	device := &usb.Device{}

	// https://github.com/torvalds/linux/blob/master/drivers/usb/gadget/legacy/zero.c
	configureDevice(device)
	configureSourceSink(device)
	configureLoopback(device)

	usb.USB1.Init()
	usb.USB1.DeviceMode()
	usb.USB1.Reset()

	// never returns
	usb.USB1.Start(device)
}
