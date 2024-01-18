// USB device mode support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package usb

import (
	"sync"
	"time"

	"github.com/usbarmory/tamago/internal/reg"
)

// DeviceMode sets the USB controller in device mode.
func (hw *USB) DeviceMode() {
	hw.Lock()
	defer hw.Unlock()

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

	// clear pending interrupts
	reg.WriteBack(hw.sts)

	// run
	reg.Set(hw.cmd, USBCMD_RS)
}

// Start waits and handles configured USB endpoints in device mode, it should
// never return. Note that isochronous endpoints are not supported.
func (hw *USB) Start(dev *Device) {
	if dev == nil {
		return
	}

	hw.Device = dev

	for {
		// check for bus reset
		if reg.Get(hw.sts, USBSTS_URI, 1) == 1 {
			// set inactive configuration
			hw.Device.ConfigurationValue = 0

			// perform controller reset procedure
			hw.Reset()
		}

		// wait for a setup packet
		if !reg.WaitFor(10*time.Millisecond, hw.setup, 0, 1, 1) {
			continue
		}

		if conf, _ := hw.handleSetup(); conf == 0 {
			continue
		}

		// stop configuration endpoints
		if hw.exit != nil {
			close(hw.exit)
			hw.wg.Wait()
		}

		// start configuration endpoints
		hw.startEndpoints()
	}
}

// ServiceInterrupts services pending endpoint transfer and bus reset events.
func (hw *USB) ServiceInterrupts() {
	defer reg.WriteBack(hw.sts)

	if hw.Device == nil {
		return
	}

	// check for bus reset
	if reg.Get(hw.sts, USBSTS_URI, 1) == 1 {
		// set inactive configuration
		hw.Device.ConfigurationValue = 0

		// perform controller reset procedure
		hw.Reset()
	}

	// process setup packets
	for reg.Read(hw.setup) != 0 {
		if conf, _ := hw.handleSetup(); conf != 0 {
			// stop configuration endpoints
			if hw.exit != nil {
				close(hw.exit)
				hw.wg.Wait()
			}

			// set interrupt event announcement
			hw.event = sync.NewCond(hw)

			// start configuration endpoints
			hw.startEndpoints()
		}
	}

	if hw.event != nil {
		// announce interrupt event to endpoints waiting transfer
		hw.event.Broadcast()
	}
}
