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
	"fmt"
	"time"
	"unsafe"

	"github.com/inversepath/tamago/imx6/internal/cache"
	"github.com/inversepath/tamago/imx6/internal/reg"
)

// Set device mode.
func (hw *usb) DeviceMode() {
	hw.Lock()
	defer hw.Unlock()

	print("imx6_usb: resetting...")
	reg.Set(hw.cmd, USBCMD_RST)
	reg.Wait(hw.cmd, USBCMD_RST, 0b1, 0)
	print("done\n")

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
	*(hw.ep) = uint32(hw.EP.addr)

	// set control endpoint
	hw.EP.set(0, IN, 64, 0, 0)
	hw.EP.set(0, OUT, 64, 0, 0)

	// Endpoint 0 is designed as a control endpoint only and does
	// not need to be configured using ENDPTCTRL0 register.
	//*(*uint32)(unsafe.Pointer(uintptr(0x021841c0))) |= (1 << 16 | 1 << 0)

	// set OTG termination
	otg := (*uint32)(unsafe.Pointer(uintptr(USB_UOG1_OTGSC)))
	reg.Set(otg, OTGSC_OT)

	// clear all pending interrupts
	*(hw.sts) = 0xffffffff

	// run
	reg.Set(hw.cmd, USBCMD_RS)

	return
}

// Wait and handle setup packets for device enumeration and configuration.
func (hw *usb) SetupHandler(dev *Device, timeout time.Duration, loop bool) (err error) {
	hw.Lock()
	defer hw.Unlock()

	hw.reset()

	for {
		setup := hw.getSetup(timeout)

		if setup == nil {
			continue
		}

		switch setup.bRequest {
		case GET_DESCRIPTOR:
			bDescriptorType := setup.wValue & 0xff
			index := setup.wValue >> 8

			switch bDescriptorType {
			case DEVICE:
				err = hw.transfer(0, dev.Descriptor.Bytes(), nil)
			case CONFIGURATION:
				var conf []byte
				if conf, err = dev.Configuration(index, setup.wLength); err == nil {
					fmt.Printf("imx6_usb: sending configuration descriptor %d (%d bytes)\n", index, setup.wLength)
					err = hw.transfer(0, conf, nil)
				}
			case STRING:
				if int(index+1) > len(dev.Strings) {
					hw.ack(0)
					err = fmt.Errorf("invalid string descriptor index %d", index)
				} else {
					if index == 0 {
						fmt.Printf("imx6_usb: sending string descriptor zero\n")
					} else {
						fmt.Printf("imx6_usb: sending string descriptor %d: \"%s\"\n", index, dev.Strings[index][2:])
					}

					err = hw.transfer(0, dev.Strings[index], nil)
				}
			case DEVICE_QUALIFIER:
				err = hw.transfer(0, dev.Qualifier.Bytes(), nil)
			default:
				hw.ack(0)
				err = fmt.Errorf("unsupported descriptor type %#x", bDescriptorType)
			}
		case SET_ADDRESS:
			addr := uint32((setup.wValue<<8)&0xff00 | (setup.wValue >> 8))
			fmt.Printf("imx6_usb: setting address %d\n", addr)

			reg.Set(hw.addr, DEVICEADDR_USBADRA)
			reg.SetN(hw.addr, DEVICEADDR_USBADR, 0x7f, addr)

			err = hw.ack(0)
		case SET_CONFIGURATION:
			conf := uint8(setup.wValue >> 8)
			fmt.Printf("imx6_usb: setting configuration value %d\n", conf)

			dev.ConfigurationValue = conf
			err = hw.ack(0)
		case GET_STATUS:
			// no meaningful status to report for now
			err = hw.transfer(0, []byte{0x00, 0x00}, nil)
		default:
			hw.ack(0)
			err = fmt.Errorf("unsupported request code: %#x", setup.bRequest)
		}

		if !loop {
			return
		}

		if err != nil {
			fmt.Printf("imx6_usb: %v\n", err)
		}
	}
}

// p3809, 56.4.6.6 Managing Transfers with Transfer Descriptors, IMX6ULLRM
func (hw *usb) transferDTD(n int, dir int, ioc bool, data []byte) (err error) {
	err = hw.EP.setDTD(n, dir, ioc, data)

	if err != nil {
		return
	}

	// TODO: clean specific cache lines instead
	cache.FlushData()

	// IN:ENDPTPRIME_PETB+n OUT:ENDPTPRIME_PERB+n
	pos := (dir * 16) + n
	// prime endpoint
	reg.Set(hw.prime, pos)

	// wait for priming completion
	reg.Wait(hw.prime, pos, 0b1, 0)
	// wait for status
	reg.WaitFor(500*time.Millisecond, &hw.EP.get(n, dir).current.token, 7, 0b1, 0)

	if status := reg.Get(&hw.EP.get(n, dir).current.token, 0, 0xff); status != 0x00 {
		err = fmt.Errorf("transfer error %x", status)
	}

	return
}

func (hw *usb) transferWait(n int, dir int) {
	// wait for transfer interrupt
	reg.Wait(hw.sts, USBSTS_UI, 0b1, 1)
	// clear interrupt
	*(hw.sts) |= 1 << USBSTS_UI

	// IN:ENDPTCOMPLETE_ETCE+n OUT:ENDPTCOMPLETE_ERCE+n
	pos := (dir * 16) + n
	// wait for completion
	reg.Wait(hw.complete, pos, 0b1, 1)

	// clear transfer completion
	*(hw.complete) |= 1 << pos
}

func (hw *usb) ack(n int) (err error) {
	err = hw.transferDTD(n, IN, true, nil)

	if err != nil {
		return
	}

	hw.transferWait(n, IN)

	return
}

func (hw *usb) transfer(n int, in []byte, out []byte) (err error) {
	err = hw.transferDTD(n, IN, false, in)

	if err != nil {
		return
	}

	err = hw.transferDTD(n, OUT, true, out)

	if err != nil {
		return
	}

	hw.transferWait(n, IN)

	return
}
