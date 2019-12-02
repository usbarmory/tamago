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
