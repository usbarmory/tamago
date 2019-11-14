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
)

type dTD struct {
	// TODO
}

type dQH struct {
	info    uint32
	current *dTD
	next    *dTD
	token   uint32

	buffer0 *[]byte
	buffer1 *[]byte
	buffer2 *[]byte
	buffer3 *[]byte
	buffer4 *[]byte

	_res uint32

	// The Set-up Buffer will be filled by hardware, note that after this
	// happens endianess needs to be adjusted with SetupData.swap().
	setup SetupData
}

// p3783, 56.4.5 Device Data Structures, IMX6ULLRM
type EndPointList struct {
	List *[MAX_ENDPOINTS * 2]dQH

	// alignment buffer
	addr uintptr
	buf  []byte
}

func (ep *EndPointList) init() {
	var err error

	ep.buf, ep.addr, err = mem.AlignedBuffer(unsafe.Sizeof(ep.List), 2048)

	if err != nil {
		panic("imx6_usb: queue head alignment error\n")
	}

	ep.List = (*[MAX_ENDPOINTS * 2]dQH)(unsafe.Pointer(ep.addr))
}

// Get endpoint queue head.
func (ep *EndPointList) Get(n int, dir int) dQH {
	// TODO: clean specific cache lines instead
	cache.FlushData()
	return ep.List[n*2+dir]
}

// Initialize endpoint queue head.
func (ep *EndPointList) Set(n int, dir int, max int, zlt int, mult int) {
	// TODO: implement and move somewhere else
	//if dir == IN {
	//	set(hw.prime, ENDPTFLUSH_PETB + n, (1 << n))
	//} else {
	//	set(hw.prime, ENDPTFLUSH_PERB + n, (1 << n))
	//}

	// p3784, 56.4.5.1 Endpoint Queue Head, IMX6ULLRM

	off := n*2 + dir

	// Mult
	reg.SetN(&ep.List[off].info, 30, 0b11, uint32(mult))
	// zlt
	reg.SetN(&ep.List[off].info, 29, 0b1, uint32(zlt))
	// Maximum Packet Length
	reg.SetN(&ep.List[off].info, 16, 0x7ff, uint32(max))

	if dir == IN {
		// interrupt on setup (ios)
		reg.SetN(&ep.List[off].info, 15, 0b1, 1)
	}

	// Total bytes
	reg.SetN(&ep.List[off].token, 16, 0xffff, 8)
	// interrupt on completion (ioc)
	reg.SetN(&ep.List[off].token, 15, 0b1, 1)
}
