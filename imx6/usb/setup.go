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
	"encoding/binary"
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

// p64, Table 46: Class-Specific Request Codes,
// USB Class Definitions for Communication Devices 1.1
const (
	SET_ETHERNET_PACKET_FILTER = 0x43
)

const (
	// p279, Table 9-5. Descriptor Types, USB Specification Revision 2.0
	DEVICE                    = 0x1
	CONFIGURATION             = 0x2
	STRING                    = 0x3
	INTERFACE                 = 0x4
	ENDPOINT                  = 0x5
	DEVICE_QUALIFIER          = 0x6
	OTHER_SPEED_CONFIGURATION = 0x7
	INTERFACE_POWER           = 0x8
)

// SetupData implements
// p276, Table 9-2. Format of Setup Data, USB Specification Revision 2.0.
type SetupData struct {
	bRequestType uint8
	bRequest     uint8
	wValue       uint16
	wIndex       uint16
	wLength      uint16
}

// swap adjusts the endianness of values written in memory by the hardware, as
// they do not match the expected one by Go.
func (s *SetupData) swap() {
	b := make([]byte, 2)

	binary.BigEndian.PutUint16(b, s.wValue)
	s.wValue = binary.LittleEndian.Uint16(b)

	binary.BigEndian.PutUint16(b, s.wIndex)
	s.wIndex = binary.LittleEndian.Uint16(b)
}

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
	// flush EP0 IN
	reg.Set(hw.flush, ENDPTFLUSH_FETB+0)
	// flush EP0 OUT
	reg.Set(hw.flush, ENDPTFLUSH_FERB+0)

	*setup = hw.EP.get(0, OUT).setup
	setup.swap()

	return
}

func (hw *usb) doSetup(dev *Device, setup *SetupData) (err error) {
	if setup == nil {
		return
	}

	switch setup.bRequest {
	case GET_STATUS:
		// no meaningful status to report for now
		log.Printf("imx6_usb: sending device status\n")
		err = hw.tx(0, false, []byte{0x00, 0x00})
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
			err = hw.tx(0, false, dev.Descriptor.Bytes())
		case CONFIGURATION:
			var conf []byte
			if conf, err = dev.Configuration(index, setup.wLength); err == nil {
				log.Printf("imx6_usb: sending configuration descriptor %d (%d bytes)\n", index, setup.wLength)
				err = hw.tx(0, false, conf)
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

				err = hw.tx(0, false, dev.Strings[index])
			}
		case DEVICE_QUALIFIER:
			log.Printf("imx6_usb: sending device qualifier\n")
			err = hw.tx(0, false, dev.Qualifier.Bytes())
		default:
			hw.stall(0, IN)
			err = fmt.Errorf("unsupported descriptor type %#x", bDescriptorType)
		}
	case GET_CONFIGURATION:
		log.Printf("imx6_usb: sending configuration value %d\n", dev.ConfigurationValue)
		err = hw.tx(0, false, []byte{dev.ConfigurationValue})
	case SET_CONFIGURATION:
		value := uint8(setup.wValue >> 8)
		log.Printf("imx6_usb: setting configuration value %d\n", value)

		dev.ConfigurationValue = value
		err = hw.ack(0)
	case GET_INTERFACE:
		log.Printf("imx6_usb: sending interface alternate setting value %d\n", dev.AlternateSetting)
		err = hw.tx(0, false, []byte{dev.AlternateSetting})
	case SET_INTERFACE:
		value := uint8(setup.wValue >> 8)
		log.Printf("imx6_usb: setting interface alternate setting value %d\n", value)

		dev.AlternateSetting = value
		err = hw.ack(0)
	case SET_ETHERNET_PACKET_FILTER:
		// no meaningful action for now
		err = hw.ack(0)
	default:
		hw.stall(0, IN)
		err = fmt.Errorf("unsupported request code: %#x", setup.bRequest)
	}

	return
}
