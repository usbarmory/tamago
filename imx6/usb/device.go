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
	"runtime"
	"time"

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
	m := *hw.mode

	// set device only controller
	reg.SetN(&m, USBMODE_CM, 0b11, USBMODE_CM_DEVICE)
	// disable setup lockout
	reg.Set(&m, USBMODE_SLOM)
	// disable stream mode
	reg.Clear(&m, USBMODE_SDIS)

	*hw.mode = m
	reg.Wait(hw.mode, USBMODE_CM, 0b11, USBMODE_CM_DEVICE)

	// set endpoint queue head
	hw.EP.init()
	*(hw.ep) = uint32(hw.EP.buf.Addr)

	// set control endpoint
	hw.EP.set(0, IN, 64, 0, 0)
	hw.EP.set(0, OUT, 64, 0, 0)

	// set OTG termination
	reg.Set(hw.otg, OTGSC_OT)

	// clear all pending interrupts
	*(hw.sts) = 0xffffffff

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
		if dev.ConfigurationValue != conf {
			// TODO: flush if ep.enabled
			runtime.Gosched()
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
