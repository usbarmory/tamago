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
	"log"

	"github.com/inversepath/tamago/imx6/internal/reg"
)

// p279, Table 9-4. Standard Request Codes, USB Specification Revision 2.0
const (
	GET_STATUS        = 0
	CLEAR_FEATURE     = 1
	SET_FEATURE       = 3
	SET_ADDRESS       = 5
	GET_DESCRIPTOR    = 6
	SET_DESCRIPTOR    = 7
	GET_CONFIGURATION = 8
	SET_CONFIGURATION = 9
	GET_INTERFACE     = 10
	SET_INTERFACE     = 11
	SYNCH_FRAME       = 12
)

// p279, Table 9-5. Descriptor Types, USB Specification Revision 2.0
const (
	DEVICE                    = 1
	CONFIGURATION             = 2
	STRING                    = 3
	INTERFACE                 = 4
	ENDPOINT                  = 5
	DEVICE_QUALIFIER          = 6
	OTHER_SPEED_CONFIGURATION = 7
	INTERFACE_POWER           = 8
)

// getSetup handles an incoming SETUP packet.
func (hw *usb) getSetup() (setup *SetupData) {
	setup = &SetupData{}

	// p3801, 56.4.6.4.2.1 Setup Phase, IMX6ULLRM

	// clear setup status
	reg.Set(hw.setup, 0)
	// set tripwire
	reg.Set(hw.cmd, USBCMD_SUTW)

	// repeat if necessary
	for reg.Get(hw.cmd, USBCMD_SUTW, 0b1) == 0 {
		log.Printf("imx6_usb: retrying setup\n")
		reg.Set(hw.cmd, USBCMD_SUTW)
	}

	// clear tripwire
	reg.Clear(hw.cmd, USBCMD_SUTW)
	// flush endpoint buffers
	*(hw.flush) = 0xffffffff

	*setup = hw.EP.get(0, OUT).setup
	setup.swap()

	return
}

func (hw *usb) doSetup(dev *Device, setup *SetupData) (err error) {
	switch setup.bRequest {
	case GET_STATUS:
		// no meaningful status to report for now
		log.Printf("imx6_usb: sending device status\n")
		err = hw.transfer(0, IN, false, []byte{0x00, 0x00})
	case SET_ADDRESS:
		addr := uint32((setup.wValue<<8)&0xff00 | (setup.wValue >> 8))
		log.Printf("imx6_usb: setting address %d\n", addr)

		reg.Set(hw.addr, DEVICEADDR_USBADRA)
		reg.SetN(hw.addr, DEVICEADDR_USBADR, 0x7f, addr)

		err = hw.ack(0)
	case GET_DESCRIPTOR:
		bDescriptorType := setup.wValue & 0xff
		index := setup.wValue >> 8

		switch bDescriptorType {
		case DEVICE:
			log.Printf("imx6_usb: sending device descriptor\n")
			err = hw.transfer(0, IN, false, dev.Descriptor.Bytes())
		case CONFIGURATION:
			var conf []byte
			if conf, err = dev.Configuration(index, setup.wLength); err == nil {
				log.Printf("imx6_usb: sending configuration descriptor %d (%d bytes)\n", index, setup.wLength)
				err = hw.transfer(0, IN, false, conf)
			}
		case STRING:
			if int(index+1) > len(dev.Strings) {
				hw.stall(0, IN)
				err = fmt.Errorf("invalid string descriptor index %d", index)
			} else {
				if index == 0 {
					log.Printf("imx6_usb: sending string descriptor zero\n")
				} else {
					log.Printf("imx6_usb: sending string descriptor %d: \"%s\"\n", index, dev.Strings[index][2:])
				}

				err = hw.transfer(0, IN, false, dev.Strings[index])
			}
		case DEVICE_QUALIFIER:
			log.Printf("imx6_usb: sending device qualifier\n")
			err = hw.transfer(0, IN, false, dev.Qualifier.Bytes())
		default:
			hw.stall(0, IN)
			err = fmt.Errorf("unsupported descriptor type %#x", bDescriptorType)
		}
	case SET_CONFIGURATION:
		value := uint8(setup.wValue >> 8)
		log.Printf("imx6_usb: setting configuration value %d\n", value)

		dev.ConfigurationValue = value
		err = hw.ack(0)
	case GET_INTERFACE:
		log.Printf("imx6_usb: sending interface alternate setting value %d\n", dev.AlternateSetting)
		err = hw.transfer(0, IN, false, []byte{dev.AlternateSetting})
	case SET_INTERFACE:
		value := uint8(setup.wValue >> 8)
		log.Printf("imx6_usb: setting interface alternate setting value %d\n", value)

		dev.AlternateSetting = value
		err = hw.ack(0)
	default:
		hw.stall(0, IN)
		err = fmt.Errorf("unsupported request code: %#x", setup.bRequest)
	}

	return
}
