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
	"encoding/binary"
	"fmt"
	"log"

	"github.com/f-secure-foundry/tamago/internal/reg"
)

// Standard request codes (p279, Table 9-4, USB2.0)
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

// Descriptor types (p279, Table 9-5, USB2.0)
const (
	DEVICE                    = 1
	CONFIGURATION             = 2
	STRING                    = 3
	INTERFACE                 = 4
	ENDPOINT                  = 5
	DEVICE_QUALIFIER          = 6
	OTHER_SPEED_CONFIGURATION = 7
	INTERFACE_POWER           = 8

	// Engineering Change Notices (ECN)
	OTG                   = 9
	DEBUG                 = 10
	INTERFACE_ASSOCIATION = 11
)

// Standard feature selectors (p280, Table 9-6, USB2.0)
const (
	ENDPOINT_HALT        = 0
	DEVICE_REMOTE_WAKEUP = 1
	TEST_MODE            = 2
)

// SetupData implements
// p276, Table 9-2. Format of Setup Data, USB2.0.
type SetupData struct {
	RequestType uint8
	Request     uint8
	Value       uint16
	Index       uint16
	Length      uint16
}

// swap adjusts the endianness of values written in memory by the hardware, as
// they do not match the expected one by Go.
func (s *SetupData) swap() {
	b := make([]byte, 2)

	binary.BigEndian.PutUint16(b, s.Value)
	s.Value = binary.LittleEndian.Uint16(b)
}

func (hw *USB) getSetup() (setup *SetupData) {
	setup = &SetupData{}

	// p3801, 56.4.6.4.2.1 Setup Phase, IMX6ULLRM

	// clear setup status
	reg.Set(hw.setup, 0)
	// set tripwire
	reg.Set(hw.cmd, USBCMD_SUTW)

	// repeat if necessary
	for reg.Get(hw.cmd, USBCMD_SUTW, 1) == 0 {
		log.Printf("imx6_usb: retrying setup")
		reg.Set(hw.cmd, USBCMD_SUTW)
	}

	// clear tripwire
	reg.Clear(hw.cmd, USBCMD_SUTW)
	// flush EP0 IN
	reg.Set(hw.flush, ENDPTFLUSH_FETB+0)
	// flush EP0 OUT
	reg.Set(hw.flush, ENDPTFLUSH_FERB+0)

	*setup = hw.getEP(0, OUT).Setup
	setup.swap()

	return
}

func (hw *USB) getDescriptor(dev *Device, setup *SetupData) (err error) {
	bDescriptorType := setup.Value & 0xff
	index := setup.Value >> 8

	switch bDescriptorType {
	case DEVICE:
		log.Printf("imx6_usb: sending device descriptor")
		err = hw.tx(0, false, trim(dev.Descriptor.Bytes(), setup.Length))
	case CONFIGURATION:
		var conf []byte
		if conf, err = dev.Configuration(index); err == nil {
			log.Printf("imx6_usb: sending configuration descriptor %d (%d bytes)", index, setup.Length)
			err = hw.tx(0, false, trim(conf, setup.Length))
		}
	case STRING:
		if int(index+1) > len(dev.Strings) {
			hw.stall(0, IN)
			err = fmt.Errorf("invalid string descriptor index %d", index)
		} else {
			if index == 0 {
				log.Printf("imx6_usb: sending string descriptor zero")
			} else {
				log.Printf("imx6_usb: sending string descriptor %d: \"%s\"", index, dev.Strings[index][2:])
			}

			err = hw.tx(0, false, trim(dev.Strings[index], setup.Length))
		}
	case DEVICE_QUALIFIER:
		log.Printf("imx6_usb: sending device qualifier")
		err = hw.tx(0, false, dev.Qualifier.Bytes())
	default:
		hw.stall(0, IN)
		err = fmt.Errorf("unsupported descriptor type %#x", bDescriptorType)
	}

	return
}

func (hw *USB) doSetup(dev *Device, setup *SetupData) (err error) {
	if setup == nil {
		return
	}

	switch setup.Request {
	case GET_STATUS:
		// no meaningful status to report for now
		log.Printf("imx6_usb: sending device status")
		err = hw.tx(0, false, []byte{0x00, 0x00})
	case CLEAR_FEATURE:
		switch setup.Value {
		case ENDPOINT_HALT:
			n := int(setup.Index & 0b1111)
			dir := int(setup.Index&0b10000000) / 0b10000000
			log.Printf("imx6_usb: EP%d.%d resetting PID", n, dir)

			hw.reset(n, dir)
			err = hw.ack(0)
		default:
			hw.stall(0, IN)
		}
	case SET_ADDRESS:
		addr := uint32((setup.Value<<8)&0xff00 | (setup.Value >> 8))
		log.Printf("imx6_usb: setting address %d", addr)

		reg.Set(hw.addr, DEVICEADDR_USBADRA)
		reg.SetN(hw.addr, DEVICEADDR_USBADR, 0x7f, addr)

		err = hw.ack(0)
	case GET_DESCRIPTOR:
		err = hw.getDescriptor(dev, setup)
	case GET_CONFIGURATION:
		log.Printf("imx6_usb: sending configuration value %d", dev.ConfigurationValue)
		err = hw.tx(0, false, []byte{dev.ConfigurationValue})
	case SET_CONFIGURATION:
		value := uint8(setup.Value >> 8)
		log.Printf("imx6_usb: setting configuration value %d", value)

		dev.ConfigurationValue = value
		err = hw.ack(0)
	case GET_INTERFACE:
		log.Printf("imx6_usb: sending interface alternate setting value %d", dev.AlternateSetting)
		err = hw.tx(0, false, []byte{dev.AlternateSetting})
	case SET_INTERFACE:
		value := uint8(setup.Value >> 8)
		log.Printf("imx6_usb: setting interface alternate setting value %d", value)

		dev.AlternateSetting = value
		err = hw.ack(0)
	case SET_ETHERNET_PACKET_FILTER:
		// no meaningful action for now
		err = hw.ack(0)
	default:
		var in []byte

		if dev.Setup == nil {
			hw.stall(0, IN)
			return fmt.Errorf("unsupported request code: %#x", setup.Request)
		}

		in, err = dev.Setup(setup)

		if err != nil {
			hw.stall(0, IN)
		} else if len(in) != 0 {
			err = hw.tx(0, false, in)
		} else {
			err = hw.ack(0)
		}
	}

	return
}

func trim(buf []byte, wLength uint16) []byte {
	if int(wLength) < len(buf) {
		buf = buf[0:wLength]
	}

	return buf
}
