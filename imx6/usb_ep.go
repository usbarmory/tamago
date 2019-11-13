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

package imx6

import (
	"unsafe"
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

type dQH [16]uint32

// p3783, 56.4.5 Device Data Structures, IMX6ULLRM
type EndPointList struct {
	List *[MAX_ENDPOINTS * 2]dQH

	// alignment buffer
	addr uintptr
	buf  []byte
}

func (ep *EndPointList) init() {
	var err error

	ep.buf, ep.addr, err = alignedBuffer(unsafe.Sizeof(ep.List), 2048)

	if err != nil {
		panic("imx6_usb: queue head alignment error\n")
	}

	ep.List = (*[MAX_ENDPOINTS * 2]dQH)(unsafe.Pointer(ep.addr))
}

// Get endpoint queue head.
func (ep *EndPointList) Get(n int, dir int) dQH {
	// TODO: clean specific cache lines instead
	v7_flush_dcache_all()
	return ep.List[n*2+dir]
}

// Initialize endpoint queue head.
func (ep *EndPointList) Set(n int, dir int, max int, zlt int, mult int) {
	if n != 0 {
		panic("imx_usb: endpoints > 0 are unsupported for now (TODO)\n")

		//fmt.Printf("imx6_usb: waiting for endpoint %d priming...", n)
		//wait(hw.prime, n, 0b1, 0)
		//fmt.Printf("done\n")
	}

	// p3784, 56.4.5.1 Endpoint Queue Head, IMX6ULLRM

	off := n*2 + dir

	// Mult
	setN(&ep.List[off][0], 30, 0b11, uint32(mult))
	// zlt
	setN(&ep.List[off][0], 29, 0b1, uint32(zlt))
	// Maximum Packet Length
	setN(&ep.List[off][0], 16, 0x7ff, uint32(max))

	if dir == IN {
		// interrupt on setup (ios)
		setN(&ep.List[off][0], 15, 0b1, 1)
	}

	// Next dTD Pointer (not terminate)
	setN(&ep.List[off][2], 0, 0b1, 0)

	// Total bytes
	setN(&ep.List[off][3], 16, 0xffff, 8)
	// interrupt on completion (ioc)
	setN(&ep.List[off][3], 15, 0b1, 1)
}
