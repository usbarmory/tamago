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
	"time"
	"unsafe"

	"github.com/inversepath/tamago/imx6/internal/cache"
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

// p276, Table 9-2. Format of Setup Data, USB Specification Revision 2.0
type SetupData struct {
	bRequestType uint8
	bRequest     uint8
	wValue       uint16
	wIndex       uint16
	wLength      uint16
}

// The endianness values written in memory by the hardware does not match the
// expected one by Go, so we have to swap multi byte values.
func (s *SetupData) swap() {
	b := make([]byte, 2)

	binary.BigEndian.PutUint16(b, s.wValue)
	s.wValue = binary.LittleEndian.Uint16(b)

	binary.BigEndian.PutUint16(b, s.wIndex)
	s.wIndex = binary.LittleEndian.Uint16(b)

	binary.BigEndian.PutUint16(b, s.wLength)
	s.wLength = binary.LittleEndian.Uint16(b)
}

// p290, Table 9-8. Standard Device Descriptor, USB Specification Revision 2.0
type DeviceDescriptor struct {
	Length            uint8
	DescriptorType    uint8
	bcdUSB            uint16
	DeviceClass       uint8
	DeviceSubClass    uint8
	DeviceProtocol    uint8
	MaxPacketSize     uint8
	VendorId          uint16
	ProductId         uint16
	Device            uint16
	Manufacturer      uint8
	Product           uint8
	SerialNumber      uint8
	NumConfigurations uint8
}

// Set default values for USB device descriptor.
func (d *DeviceDescriptor) SetDefaults() {
	d.Length = uint8(unsafe.Sizeof(DeviceDescriptor{}))
	d.DescriptorType = DEVICE
	// USB 2.0
	d.bcdUSB = 0x0200
	// maximum packet size for EP0
	d.MaxPacketSize = 64
	// http://pid.codes/1209/2702/
	d.VendorId = 0x1209
	d.ProductId = 0x2702
	d.NumConfigurations = 1
}

// p293, Table 9-10. Standard Configuration Descriptor, USB Specification Revision 2.0
type ConfigurationDescriptor struct {
	Length             uint8
	DescriptorType     uint8
	TotalLength        uint16
	NumInterfaces      uint8
	ConfigurationValue uint8
	Configuration      uint8
	Attributes         uint8
	MaxPower           uint8
}

// Set default values for USB configuration descriptor.
func (d *ConfigurationDescriptor) SetDefaults() {
	d.Length = uint8(unsafe.Sizeof(ConfigurationDescriptor{}))
	d.DescriptorType = CONFIGURATION
	d.NumInterfaces = 1
	d.ConfigurationValue = 1
	d.Attributes = 0xc0
	d.MaxPower = 250
}

// p296, Table 9-12. Standard Interface Descriptor, USB Specification Revision 2.0
type InterfaceDescriptor struct {
	Length            uint8
	DescriptorType    uint8
	InterfaceNumber   uint8
	AlternateSetting  uint8
	NumEndpoints      uint8
	InterfaceClass    uint8
	InterfaceSubClass uint8
	InterfaceProtocol uint8
	Interface         uint8
}

// Set default values for USB interface descriptor.
func (d *InterfaceDescriptor) SetDefaults() {
	d.Length = uint8(unsafe.Sizeof(InterfaceDescriptor{}))
	d.DescriptorType = INTERFACE
	d.NumEndpoints = 1
}

// p297, Table 9-13. Standard Endpoint Descriptor, USB Specification Revision 2.0
type EndpointDescriptor struct {
	Length          uint8
	DescriptorType  uint8
	EndpointAddress uint8
	Attributes      uint8
	MaxPacketSize   uint16
	Interval        uint8
}

// Set default values for USB endpoint descriptor.
func (d *EndpointDescriptor) SetDefaults() {
	d.Length = uint8(unsafe.Sizeof(EndpointDescriptor{}))
	d.DescriptorType = ENDPOINT
	// EP1 IN
	d.EndpointAddress = 0x81
}

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

// Perform device enumeration.
func (hw *usb) DeviceEnumeration(desc *DeviceDescriptor) (err error) {
	hw.Lock()
	defer hw.Unlock()

	hw.reset()

	for {
		if err != nil {
			return
		}

		setup := hw.getSetup()
		fmt.Printf("imx6_usb: got setup packet %+v\n", setup)

		switch setup.bRequest {
		case GET_DESCRIPTOR:
			switch setup.wValue {
			case DEVICE:
				err = hw.transfer(0, desc, nil)
			default:
				return fmt.Errorf("unsupported descriptor type %#x", setup.wValue)
			}
		case SET_ADDRESS:
			addr := uint32((setup.wValue<<8)&0xff00 | (setup.wValue >> 8))
			fmt.Printf("imx6_usb: set address %d\n", addr)

			reg.Set(hw.addr, DEVICEADDR_USBADRA)
			reg.SetN(hw.addr, DEVICEADDR_USBADR, 0x7f, addr)

			err = hw.ack(0)
		default:
			return fmt.Errorf("unsupported request code: %#x", setup.bRequest)
		}
	}
}

// p3809, 56.4.6.6 Managing Transfers with Transfer Descriptors, IMX6ULLRM
func (hw *usb) transferDTD(n int, dir int, ioc bool, buf interface{}) (err error) {
	err = hw.EP.setDTD(n, dir, ioc, buf)

	if err != nil {
		return
	}

	// TODO: clean specific cache lines instead
	cache.FlushData()

	// IN:ENDPTPRIME_PETB+n OUT:ENDPTPRIME_PERB+n
	pos := (dir * 16) + n

	fmt.Printf("imx6_usb: priming endpoint %d.%d transfer...", n, dir)
	reg.Set(hw.prime, pos)

	print("waiting completion...")
	reg.Wait(hw.prime, pos, 0b1, 0)
	print("waiting status...")
	reg.WaitFor(500*time.Millisecond, &hw.EP.get(n, dir).current.token, 7, 0b1, 0)
	print("...")

	if status := reg.Get(&hw.EP.get(n, dir).current.token, 0, 0xff); status != 0x00 {
		print("error\n")
		return fmt.Errorf("transfer error %x", status)
	}
	print("done\n")

	return
}

func (hw *usb) transferWait(n int, dir int) {
	print("imx6_usb: waiting for transfer interrupt...")
	reg.Wait(hw.sts, USBSTS_UI, 0b1, 1)
	print("done\n")
	// clear interrupt
	*(hw.sts) |= 1 << USBSTS_UI

	// IN:ENDPTCOMPLETE_ETCE+n OUT:ENDPTCOMPLETE_ERCE+n
	pos := (dir * 16) + n

	print("imx6_usb: waiting for endpoint transfer completion...")
	reg.Wait(hw.complete, pos, 0b1, 1)
	print("done\n")
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

func (hw *usb) transfer(n int, in interface{}, out interface{}) (err error) {
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
