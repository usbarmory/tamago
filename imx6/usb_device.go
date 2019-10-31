// USB device mode
// https://github.com/inversepath/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build tamago,arm

package imx6

import (
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

// p290, Table 9-8. Standard Device Descriptor, USB Specification Revision 2.0
type DeviceDescriptor struct {
	bLength            uint8
	bDescriptorType    uint8
	bcdUSB             uint16
	bDeviceClass       uint8
	bDeviceSubClass    uint8
	bDeviceProtocol    uint8
	bMaxPacketSize     uint8
	idVendor           uint16
	idProduct          uint16
	bcdDevice          uint16
	iManufacturer      uint8
	iProduct           uint8
	iSerialNumber      uint8
	bNumConfigurations uint8
}

// p293, Table 9-10. Standard Configuration Descriptor, USB Specification Revision 2.0
type ConfigurationDescriptor struct {
	bLength             uint8
	bDescriptorType     uint8
	wTotalLength        uint16
	bNumInterfaces      uint8
	bConfigurationValue uint8
	iConfiguration      uint8
	bmAttributes        uint8
	maxPower            uint8
}

// p296, Table 9-12. Standard Interface Descriptor, USB Specification Revision 2.0
type InterfaceDescriptor struct {
	bLength            uint8
	bDescriptorType    uint8
	bInterfaceNumber   uint8
	bAlternateSetting  uint8
	bNumEndpoints      uint8
	bInterfaceClass    uint8
	bInterfaceSubClass uint8
	bInterfaceProtocol uint8
	iInterface         uint8
}

// p297, Table 9-13. Standard Endpoint Descriptor, USB Specification Revision 2.0
type EndpointDescriptor struct {
	bLength          uint8
	bDescriptorType  uint8
	bEndpointAddress uint8
	bmAttributes     uint8
	wMaxPacketSize   uint16
	bInterval        uint8
}

// Set device mode.
func (hw *usb) DeviceMode() {
	hw.Lock()
	defer hw.Unlock()

	print("imx6_usb: resetting USB1...")
	set(hw.cmd, USBCMD_RST)
	wait(hw.cmd, USBCMD_RST, 0b1, 0)
	print("done\n")

	// p3872, 56.6.33 USB Device Mode (USB_nUSBMODE), IMX6ULLRM)
	mode := (*uint32)(unsafe.Pointer(uintptr(USB_UOG1_USBMODE)))
	m := *mode

	// set device only controller
	setN(&m, USBMODE_CM, 0b11, USBMODE_CM_DEVICE)
	// disable setup lockout
	set(&m, USBMODE_SLOM)
	// disable stream mode
	clear(&m, USBMODE_SDIS)

	*mode = m
	wait(mode, USBMODE_CM, 0b11, USBMODE_CM_DEVICE)

	// set endpoint queue head
	hw.EP.init()
	*(hw.ep) = uint32(hw.EP.addr)

	// set control endpoint
	hw.EP.Set(0, IN, 64, 0, 0)
	hw.EP.Set(0, OUT, 64, 0, 0)

	// Endpoint 0 is designed as a control endpoint only and does
	// not need to be configured using ENDPTCTRL0 register.
	//*(*uint32)(unsafe.Pointer(uintptr(0x021841c0))) |= (1 << 16 | 1 << 0)

	// set OTG termination
	otg := (*uint32)(unsafe.Pointer(uintptr(USB_UOG1_OTGSC)))
	set(otg, OTGSC_OT)

	// clear all pending interrupts
	*(hw.sts) = 0xffffffff

	// run
	set(hw.cmd, USBCMD_RS)

	// start control transaction
	hw.controlTransaction()

	return
}
