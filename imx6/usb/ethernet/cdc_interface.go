// Ethernet over USB driver
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package ethernet

import (
	"strings"

	"github.com/f-secure-foundry/tamago/imx6/usb"
)

// Build a CDC control interface.
func (eth *NIC) buildControlInterface(device *usb.Device) (iface *usb.InterfaceDescriptor) {
	iface = &usb.InterfaceDescriptor{}
	iface.SetDefaults()

	iface.NumEndpoints = 1
	iface.InterfaceClass = 2
	iface.InterfaceSubClass = 6

	iInterface, _ := device.AddString(`CDC Ethernet Control Model (ECM)`)
	iface.Interface = iInterface

	// Set IAD to be inserted before first interface, to support multiple
	// functions in this same configuration.
	iface.IAD = &usb.InterfaceAssociationDescriptor{}
	iface.IAD.SetDefaults()
	// alternate settings do not count
	iface.IAD.InterfaceCount = 1
	iface.IAD.FunctionClass = iface.InterfaceClass
	iface.IAD.FunctionSubClass = iface.InterfaceSubClass

	iFunction, _ := device.AddString(`CDC`)
	iface.IAD.Function = iFunction

	header := &usb.CDCHeaderDescriptor{}
	header.SetDefaults()

	iface.ClassDescriptors = append(iface.ClassDescriptors, header.Bytes())

	union := &usb.CDCUnionDescriptor{}
	union.SetDefaults()

	iface.ClassDescriptors = append(iface.ClassDescriptors, union.Bytes())

	ethernet := &usb.CDCEthernetDescriptor{}
	ethernet.SetDefaults()

	iMacAddress, _ := device.AddString(strings.ReplaceAll(eth.Host.String(), ":", ""))
	ethernet.MacAddress = iMacAddress

	iface.ClassDescriptors = append(iface.ClassDescriptors, ethernet.Bytes())

	ep2IN := &usb.EndpointDescriptor{}
	ep2IN.SetDefaults()
	ep2IN.EndpointAddress = 0x82
	ep2IN.Attributes = 3
	ep2IN.MaxPacketSize = 16
	ep2IN.Interval = 9
	ep2IN.Function = eth.Control

	iface.Endpoints = append(iface.Endpoints, ep2IN)

	return
}

// Build a CDC data interface.
func (eth *NIC) buildDataInterface(device *usb.Device) (iface *usb.InterfaceDescriptor) {
	iface = &usb.InterfaceDescriptor{}
	iface.SetDefaults()

	// ECM requires the use of "alternate settings" for its data interface
	iface.AlternateSetting = 1
	iface.NumEndpoints = 2
	iface.InterfaceClass = 10

	iInterface, _ := device.AddString(`CDC Data`)
	iface.Interface = iInterface

	ep1IN := &usb.EndpointDescriptor{}
	ep1IN.SetDefaults()
	ep1IN.EndpointAddress = 0x81
	ep1IN.Attributes = 2
	ep1IN.Function = eth.Tx

	iface.Endpoints = append(iface.Endpoints, ep1IN)

	ep1OUT := &usb.EndpointDescriptor{}
	ep1OUT.SetDefaults()
	ep1OUT.EndpointAddress = 0x01
	ep1OUT.Attributes = 2
	ep1OUT.Function = eth.Rx

	iface.Endpoints = append(iface.Endpoints, ep1OUT)

	return
}
