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

func TestUSB() {
	usb.USB1.Init()
	usb.USB1.DeviceMode()

	deviceDescriptor := &usb.DeviceDescriptor{}
	deviceDescriptor.SetDefaults()

	configurationDescriptor := &usb.ConfigurationDescriptor{}
	configurationDescriptor.SetDefaults()

	interfaceDescriptor := &usb.InterfaceDescriptor{}
	interfaceDescriptor.SetDefaults()

	endpointDescriptor := &usb.EndpointDescriptor{}
	endpointDescriptor.SetDefaults()

	// TODO: for now this leads to `no configurations` enumeration error on
	// host side, after an address is being successfully assigned.
	deviceDescriptor.NumConfigurations = 0

	err := usb.USB1.DeviceEnumeration(deviceDescriptor)

	// TODO
	//err := usb.USB1.DeviceEnumeration(deviceDescriptor,
	//	[]usb.ConfigurationDescriptor{configurationDescriptor},
	//	[]usb.InterfaceDescriptor{interfaceDescriptor},
	//	[]usb.EndpointDescriptor{endpointDescriptor}
	//)

	if err != nil {
		fmt.Printf("imx6_usb: enumeration error, %v\n", err)
	}
}
