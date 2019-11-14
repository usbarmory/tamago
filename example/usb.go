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

	device := &usb.DeviceDescriptor{}
	err := usb.USB1.DeviceEnumeration(device)

	if err != nil {
		fmt.Printf("imx6_usb: enumeration error, %v\n", err)
	}
}
