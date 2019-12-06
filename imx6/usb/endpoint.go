// NXP USBOH3USBO2 / USBPHY driver
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
	"errors"
	"fmt"
	"log"
	"unsafe"

	"github.com/inversepath/tamago/imx6/internal/cache"
	"github.com/inversepath/tamago/imx6/internal/mem"
	"github.com/inversepath/tamago/imx6/internal/reg"
)

const (
	// The USB OTG device controller hardware supports up to 8 endpoint
	// numbers.
	MAX_ENDPOINTS = 8

	// Host -> Device
	OUT = 0
	// Device -> Host
	IN = 1

	// Transfer Type
	CONTROL     = 0
	ISOCHRONOUS = 1
	BULK        = 2
	INTERRUPT   = 3

	// p3787, 56.4.5.2 Endpoint Transfer Descriptor (dTD), IMX6ULLRM
	DTD_PAGES     = 5
	DTD_PAGE_SIZE = 4096
)

// dTD implements
// p3787, 56.4.5.2 Endpoint Transfer Descriptor (dTD), IMX6ULLRM.
type dTD struct {
	next   *dTD
	token  uint32
	buffer [5]uintptr

	buf   mem.AlignmentBuffer
	pages mem.AlignmentBuffer
}

// dQH implements
// p3784, 56.4.5.1 Endpoint Queue Head (dQH), IMX6ULLRM.
type dQH struct {
	info    uint32
	current *dTD
	next    *dTD
	token   uint32
	buffer  [5]uintptr

	_res uint32

	// The Set-up Buffer will be filled by hardware, note that after this
	// happens endianess needs to be adjusted with SetupData.swap().
	setup SetupData

	// We align only the first queue entry, so we need this gap to maintain
	// 64-byte boundaries.
	_align [4]uint32
}

// EndPointList implements
// p3783, 56.4.5 Device Data Structures, IMX6ULLRM.
type EndPointList struct {
	List *[MAX_ENDPOINTS * 2]dQH

	buf mem.AlignmentBuffer
}

func (ep *EndPointList) init() {
	ep.buf = mem.AlignmentBuffer{}
	ep.buf.Init(unsafe.Sizeof(ep.List), 2048)

	ep.List = (*[MAX_ENDPOINTS * 2]dQH)(unsafe.Pointer(ep.buf.Addr))
}

func (ep *EndPointList) get(n int, dir int) dQH {
	// TODO: clean specific cache lines instead
	cache.FlushData()
	return ep.List[n*2+dir]
}

// set configures a queue head as described in
// p3784, 56.4.5.1 Endpoint Queue Head, IMX6ULLRM.
func (ep *EndPointList) set(n int, dir int, max int, zlt int, mult int) {
	off := n*2 + dir

	// Mult
	reg.SetN(&ep.List[off].info, 30, 0b11, uint32(mult))
	// zlt
	reg.SetN(&ep.List[off].info, 29, 0b1, uint32(zlt))
	// Maximum Packet Length
	reg.SetN(&ep.List[off].info, 16, 0x7ff, uint32(max))

	if dir == IN {
		// interrupt on setup (ios)
		reg.Set(&ep.List[off].info, 15)
	}

	// Total bytes
	reg.SetN(&ep.List[off].token, 16, 0xffff, 8)
	// interrupt on completion (ioc)
	reg.Set(&ep.List[off].token, 15)
	// multiplier override (MultO)
	reg.SetN(&ep.List[off].token, 10, 0b11, 0)
}

// setDTD configures an endpoint transfer descriptor as described in
// p3787, 56.4.5.2 Endpoint Transfer Descriptor (dTD), IMX6ULLRM.
func (ep *EndPointList) setDTD(n int, dir int, ioc bool, data []byte) (err error) {
	var dtd *dTD
	size := len(data)

	if size > DTD_PAGES*DTD_PAGE_SIZE {
		return errors.New("unsupported transfer size")
	}

	// re-use existing buffer if present
	if dtd = ep.List[n*2+dir].current; dtd == nil {
		dtd = ep.List[n*2+dir].next
	}

	if dtd == nil {
		dtdBuf := mem.AlignmentBuffer{}
		dtdBuf.Init(unsafe.Sizeof(dTD{}), 32)

		dtd = (*dTD)(unsafe.Pointer(dtdBuf.Addr))
		dtd.buf = dtdBuf
	}

	// p3809, 56.4.6.6.2 Building a Transfer Descriptor, IMX6ULLRM

	// invalidate next pointer
	dtd.next = (*dTD)(unsafe.Pointer(uintptr(1)))

	// interrupt on completion (ioc)
	if ioc {
		reg.Set(&dtd.token, 15)
	} else {
		reg.Clear(&dtd.token, 15)
	}

	// multiplier override (MultO)
	reg.SetN(&dtd.token, 10, 0b11, 0)
	// active status
	reg.Set(&dtd.token, 7)

	// re-use existing buffer if present
	if dtd.pages.Addr == 0 {
		dtd.pages = mem.AlignmentBuffer{}
		dtd.pages.Init(DTD_PAGE_SIZE*DTD_PAGES, DTD_PAGE_SIZE)
	}

	mem.Copy(dtd.pages, data)

	// total bytes
	reg.SetN(&dtd.token, 16, 0xffff, uint32(size))

	for n := 0; n < DTD_PAGES; n++ {
		dtd.buffer[n] = dtd.pages.Addr + uintptr(DTD_PAGE_SIZE*n)
	}

	ep.List[n*2+dir].next = dtd

	return
}

// transferDTD manages a transfer using transfer descriptors
// (p3809, 56.4.6.6 Managing Transfers with Transfer Descriptors, IMX6ULLRM).
func (hw *usb) transferDTD(n int, dir int, ioc bool, data []byte) (err error) {
	err = hw.EP.setDTD(n, dir, ioc, data)

	if err != nil {
		return
	}

	// TODO: clean specific cache lines instead
	cache.FlushData()

	// IN:ENDPTPRIME_PETB+n OUT:ENDPTPRIME_PERB+n
	pos := (dir * 16) + n
	// prime endpoint (TODO: we can do it once when the descriptor is added to dQH)
	reg.Set(hw.prime, pos)
	// wait for priming completion
	reg.Wait(hw.prime, pos, 0b1, 0)

	// wait for status
	reg.Wait(&hw.EP.get(n, dir).current.token, 7, 0b1, 0)

	if status := reg.Get(&hw.EP.get(n, dir).current.token, 0, 0xff); status != 0x00 {
		err = fmt.Errorf("transfer error %x", status)
	}

	return
}

func (hw *usb) transferWait(n int, dir int) (err error) {
	// TODO: interrupt check should be probably moved elsewhere and use to
	// drive all handlers

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

	return
}

func (hw *usb) transfer(n int, dir int, ioc bool, data []byte) (err error) {
	err = hw.transferDTD(n, dir, ioc, data)

	if err != nil {
		return
	}

	// acknowledge completion
	if dir == IN {
		err = hw.transferDTD(n, OUT, true, []byte{})

		if err != nil {
			return
		}
	}

	return hw.transferWait(n, dir)
}

func (hw *usb) ack(n int) (err error) {
	err = hw.transferDTD(n, IN, true, nil)

	if err != nil {
		return
	}

	return hw.transferWait(n, IN)
}

func (hw *usb) stall(n int, dir int) {
	ctrl := (*uint32)(unsafe.Pointer(uintptr(USB_UOG1_ENDPTCTRL + uint32(4*n))))

	if dir == IN {
		reg.Set(ctrl, ENDPTCTRL_TXS)
	} else {
		reg.Set(ctrl, ENDPTCTRL_RXS)
	}
}

func (hw *usb) enable(n int, dir int, transferType int) {
	log.Printf("imx6_usb: enabling EP%d.%d\n", n, dir)

	ctrl := (*uint32)(unsafe.Pointer(uintptr(USB_UOG1_ENDPTCTRL + uint32(4*n))))
	c := *ctrl

	if dir == IN {
		reg.Set(&c, ENDPTCTRL_TXE)
		reg.SetN(&c, ENDPTCTRL_TXT, 0b11, uint32(transferType))
	} else {
		reg.Set(&c, ENDPTCTRL_RXE)
		reg.SetN(&c, ENDPTCTRL_RXT, 0b11, uint32(transferType))
	}

	*ctrl = c
}
