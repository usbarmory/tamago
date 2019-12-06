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

package usb

import (
	"log"
	"time"
	"unsafe"

	"github.com/inversepath/tamago/imx6/internal/reg"
)

// DeviceMode sets the USB controller in device mode.
func (hw *usb) DeviceMode() {
	hw.Lock()
	defer hw.Unlock()

	log.Printf("imx6_usb: resetting\n")
	reg.Set(hw.cmd, USBCMD_RST)
	reg.Wait(hw.cmd, USBCMD_RST, 0b1, 0)

	// p3872, 56.6.33 USB Device Mode (USB_nUSBMODE), IMX6ULLRM)
	mode := (*uint32)(unsafe.Pointer(uintptr(USB_UOG1_USBMODE)))
	m := *mode

	// set device only controller
	reg.SetN(&m, USBMODE_CM, 0b11, USBMODE_CM_DEVICE)
	// disable setup lockout
	reg.Set(&m, USBMODE_SLOM)
	// disable stream mode
	reg.Clear(&m, USBMODE_SDIS)

	*mode = m
	reg.Wait(mode, USBMODE_CM, 0b11, USBMODE_CM_DEVICE)

	// set endpoint queue head
	hw.EP.init()
	*(hw.ep) = uint32(hw.EP.buf.Addr)

	// set control endpoint
	hw.EP.set(0, IN, 64, 0, 0)
	hw.EP.set(0, OUT, 64, 0, 0)

	// set OTG termination
	otg := (*uint32)(unsafe.Pointer(uintptr(USB_UOG1_OTGSC)))
	reg.Set(otg, OTGSC_OT)

	// clear all pending interrupts
	*(hw.sts) = 0xffffffff

	// run
	reg.Set(hw.cmd, USBCMD_RS)

	return
}

// SetupHandler waits and handles USB SETUP packets on endpoint 0 for device
// enumeration and configuration.
func (hw *usb) SetupHandler(dev *Device) {
	for {
		if !reg.WaitFor(5*time.Second, hw.setup, 0, 0b1, 1) {
			continue
		}

		setup := hw.getSetup()

		if setup == nil {
			continue
		}

		if err := hw.doSetup(dev, setup); err != nil {
			log.Printf("imx6_usb: setup error, %v\n", err)
		}
	}
}

// EndpointHandler waits and handles data transfers on a non-control endpoint.
func (hw *usb) EndpointHandler(ep *EndpointDescriptor) {
	ep.Lock()
	defer ep.Unlock()

	n := ep.Number()
	dir := ep.Direction()

	if !ep.enabled {
		hw.EP.set(n, dir, int(ep.MaxPacketSize), 0, 0)
		hw.enable(n, dir, ep.TransferType())
		ep.enabled = true
	}

	for {
		var data []byte
		var err error

		if ep.Function != nil {
			data, err = ep.Function(ep.MaxPacketSize)

			if err != nil {
				return
			}
		}

		err = hw.EP.setDTD(n, dir, true, data)

		if err != nil {
			log.Printf("imx6_usb: EP%d.%d dTD error, %v\n", n, dir, err)
			continue
		}

		err = hw.transfer(n, dir, true, data)

		if err != nil {
			log.Printf("imx6_usb: EP%d.%d transfer error, %v\n", n, dir, err)
		}
	}
}
