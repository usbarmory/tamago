// USB device mode

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
	"log"
	"runtime"
	"time"

	"github.com/f-secure-foundry/tamago/imx6/internal/reg"
)

// DeviceMode sets the USB controller in device mode.
func (hw *usb) DeviceMode() {
	hw.Lock()
	defer hw.Unlock()

	log.Printf("imx6_usb: resetting\n")
	reg.Set(hw.cmd, USBCMD_RST)
	reg.Wait(hw.cmd, USBCMD_RST, 0b1, 0)

	// p3872, 56.6.33 USB Device Mode (USB_nUSBMODE), IMX6ULLRM)
	m := reg.Read(hw.mode)

	// set device only controller
	m = (m & ^uint32(0b11<<USBMODE_CM)) | (USBMODE_CM_DEVICE << USBMODE_CM)
	// disable setup lockout
	m |= (1 << USBMODE_SLOM)
	// disable stream mode
	m &^= (1 << USBMODE_SDIS)

	reg.Write(hw.mode, m)
	reg.Wait(hw.mode, USBMODE_CM, 0b11, USBMODE_CM_DEVICE)

	// set endpoint queue head
	hw.EP.init()
	reg.Write(hw.ep, hw.EP.buf.Addr())

	// set control endpoint
	hw.EP.set(0, IN, 64, 0, 0)
	hw.EP.set(0, OUT, 64, 0, 0)

	// set OTG termination
	reg.Set(hw.otg, OTGSC_OT)

	// clear all pending interrupts
	reg.Write(hw.sts, 0xffffffff)

	// run
	reg.Set(hw.cmd, USBCMD_RS)

	return
}

// Start waits and handles configured USB endpoints, it should never return.
//
// Current limitations:
//   * bus reset after initial setup are not handled
//   * only control/bulk/interrupt endpoints are supported (e.g. no isochronous support)
func (hw *usb) Start(dev *Device) {
	for _, conf := range dev.Configurations {
		for _, iface := range conf.Interfaces {
			for _, ep := range iface.Endpoints {
				go func(ep *EndpointDescriptor, conf uint8) {
					hw.endpointHandler(dev, ep, conf)
				}(ep, conf.ConfigurationValue)
			}
		}
	}

	hw.setupHandler(dev)
}

func (hw *usb) setupHandler(dev *Device) {
	for {
		if !reg.WaitFor(10*time.Millisecond, hw.setup, 0, 0b1, 1) {
			continue
		}

		setup := hw.getSetup()

		if err := hw.doSetup(dev, setup); err != nil {
			log.Printf("imx6_usb: setup error, %v\n", err)
		}
	}
}

func (hw *usb) endpointHandler(dev *Device, ep *EndpointDescriptor, conf uint8) {
	var err error
	var out []byte
	var in []byte

	if ep.Function == nil {
		return
	}

	ep.Lock()
	defer ep.Unlock()

	n := ep.Number()
	dir := ep.Direction()

	for {
		runtime.Gosched()

		if dev.ConfigurationValue != conf {
			// TODO: flush if ep.enabled
			continue
		}

		if !ep.enabled {
			hw.EP.set(n, dir, int(ep.MaxPacketSize), 1, 0)
			hw.enable(n, dir, ep.TransferType())
			ep.enabled = true
		}

		if dir == OUT {
			out, err = hw.rx(n, true)

			if err == nil {
				_, err = ep.Function(out, err)
			}
		} else {
			in, err = ep.Function(out, err)

			if err == nil {
				err = hw.tx(n, true, in)
			}
		}

		if err != nil {
			log.Printf("imx6_usb: EP%d.%d transfer error, %v\n", n, dir, err)
		}
	}
}
