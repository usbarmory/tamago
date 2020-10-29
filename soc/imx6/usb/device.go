// USB device mode support
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package usb

import (
	"log"
	"runtime"
	"time"

	"github.com/f-secure-foundry/tamago/internal/reg"
)

// DeviceMode sets the USB controller in device mode.
func (hw *USB) DeviceMode() {
	hw.Lock()
	defer hw.Unlock()

	log.Printf("imx6_usb: resetting")
	reg.Set(hw.cmd, USBCMD_RST)
	reg.Wait(hw.cmd, USBCMD_RST, 1, 0)

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

	// initialize endpoint queue head list
	hw.initQH()
	// set control endpoint
	hw.set(0, IN, 64, true, 0)
	hw.set(0, OUT, 64, true, 0)

	// set OTG termination
	reg.Set(hw.otg, OTGSC_OT)

	// clear all pending interrupts
	reg.Write(hw.sts, 0xffffffff)

	// run
	reg.Set(hw.cmd, USBCMD_RS)
}

// Start waits and handles configured USB endpoints, it should never return.
//
// Current limitations:
//   * bus reset after initial setup are not handled
//   * only control/bulk/interrupt endpoints are supported (e.g. no isochronous support)
func (hw *USB) Start(dev *Device) {
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

func (hw *USB) setupHandler(dev *Device) {
	for {
		if !reg.WaitFor(10*time.Millisecond, hw.setup, 0, 1, 1) {
			continue
		}

		setup := hw.getSetup()

		if err := hw.doSetup(dev, setup); err != nil {
			log.Printf("imx6_usb: setup error, %v", err)
		}
	}
}

func (hw *USB) endpointHandler(dev *Device, ep *EndpointDescriptor, conf uint8) {
	var err error
	var buf []byte
	var res []byte

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
			if ep.enabled {
				reg.Set(hw.flush, (dir*16)+n)
				ep.enabled = false
			}

			continue
		}

		if !ep.enabled {
			hw.set(n, dir, int(ep.MaxPacketSize), ep.Zero, 0)
			hw.enable(n, dir, ep.TransferType())
			ep.enabled = true
		}

		if dir == OUT {
			buf, err = hw.rx(n, false, res)

			if err == nil && len(buf) != 0 {
				res, err = ep.Function(buf, err)
			}
		} else {
			res, err = ep.Function(nil, err)

			if err == nil && len(res) != 0 {
				err = hw.tx(n, false, res)
			}
		}

		if err != nil {
			reg.Set(hw.flush, (dir*16)+n)
			log.Printf("imx6_usb: EP%d.%d transfer error, %v", n, dir, err)
		}
	}
}
