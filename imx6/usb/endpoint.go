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

	// We align only the first queue entry, so we need a 4*uint32 gap to
	// maintain 64-byte boundaries, we re-use this space for queue
	// pointers.
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

// get returns the Endpoint Queue Head (dQH)
func (ep *EndPointList) get(n int, dir int) dQH {
	// TODO: clean specific cache lines instead
	cache.FlushData()
	return ep.List[n*2+dir]
}

// max returns the endpoint Maximum Packet Length
func (ep *EndPointList) max(n int, dir int) int {
	return int(ep.List[n*2+dir].info>>16) & 0x7ff
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

	if n == 0 && dir == IN {
		// interrupt on setup (ios)
		reg.Set(&ep.List[off].info, 15)
	}

	// Total bytes
	reg.SetN(&ep.List[off].token, 16, 0xffff, 0)
	// interrupt on completion (ioc)
	reg.Set(&ep.List[off].token, 15)
	// multiplier override (MultO)
	reg.SetN(&ep.List[off].token, 10, 0b11, 0)
}

// addDTD configures an endpoint transfer descriptor as described in
// p3787, 56.4.5.2 Endpoint Transfer Descriptor (dTD), IMX6ULLRM.
func buildDTD(n int, dir int, ioc bool, data []byte) (dtd *dTD, err error) {
	size := len(data)

	if size > DTD_PAGES*DTD_PAGE_SIZE {
		return nil, errors.New("unsupported transfer size")
	}

	dtdBuf := mem.AlignmentBuffer{}
	dtdBuf.Init(unsafe.Sizeof(dTD{}), 32)

	dtd = (*dTD)(unsafe.Pointer(dtdBuf.Addr))
	dtd.buf = dtdBuf

	// p3809, 56.4.6.6.2 Building a Transfer Descriptor, IMX6ULLRM

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

	dtd.pages = mem.AlignmentBuffer{}
	dtd.pages.Init(DTD_PAGE_SIZE*DTD_PAGES, DTD_PAGE_SIZE)
	mem.Copy(dtd.pages, data)

	// total bytes
	reg.SetN(&dtd.token, 16, 0xffff, uint32(size))

	for n := 0; n < DTD_PAGES; n++ {
		dtd.buffer[n] = dtd.pages.Addr + uintptr(DTD_PAGE_SIZE*n)
	}

	// invalidate next pointer
	dtd.next = (*dTD)(unsafe.Pointer(uintptr(1)))

	return
}

// transferDTD manages a transfer using transfer descriptors (dTDs) as
// described in p3809, 56.4.6.6 Managing Transfers with Transfer Descriptors,
// IMX6ULLRM.
func (hw *usb) transferDTD(n int, dir int, ioc bool, data []byte) (err error) {
	var dtds []*dTD
	dtdLength := len(data)

	if n != 0 {
		// On non-control endpoints, For simplicity, each dTD is
		// configured to transfer one packet
		// (dTD.TotalBytes == dQH.MaxPacketLength).
		max := hw.EP.max(n, dir)

		if len(data) > max {
			dtdLength = max
		}
	}

	dtd, err := buildDTD(n, dir, ioc, data[0:dtdLength])

	if err != nil {
		return
	}

	off := n*2 + dir
	// set dQH head pointer
	hw.EP.List[off].next = dtd
	// reset dQH status
	reg.SetN(&hw.EP.List[off].token, 0, 0xff, 0)

	dtds = append(dtds, dtd)

	for i := dtdLength; i < len(data); i += dtdLength {
		size := i + dtdLength

		if size > len(data) {
			size = len(data)
		}

		next, err := buildDTD(n, dir, ioc, data[i:size])

		if err != nil {
			return err
		}

		dtd.next = next
		dtd = next
		dtds = append(dtds, next)
	}

	// hw.prime IN:ENDPTPRIME_PETB+n    OUT:ENDPTPRIME_PERB+n
	// hw.pos   IN:ENDPTCOMPLETE_ETCE+n OUT:ENDPTCOMPLETE_ERCE+n
	pos := (dir * 16) + n

	// prime endpoint
	reg.Write(hw.prime, 1<<pos)
	// wait for priming completion
	reg.Wait(hw.prime, pos, 0b1, 0)

	// wait for completion
	reg.Wait(hw.complete, pos, 0b1, 1)
	// clear completion
	reg.Write(hw.complete, 1<<pos)

	// OPTIMIZE: waiting for each dTD status before re-priming and
	// re-creating the DMA transfer is not optimal, but for simplicity we
	// keep things like this for now as performance remains good.
	//
	// In the future return as soon as completion is detected and worry
	// about checking status only on the next iteration, which can be used
	// to fill the buffer so that at least one transfer is always queued.

	for i, dtd := range dtds {
		reg.Wait(&dtd.token, 7, 0b1, 0)

		if status := (dtd.token & 0xff); status != 0x00 {
			return fmt.Errorf("error status for dTD #%d, %x", i, status)
		}
	}

	return
}

func (hw *usb) transfer(n int, dir int, ioc bool, data []byte) (err error) {
	err = hw.transferDTD(n, dir, ioc, data)

	if err != nil {
		return
	}

	// p3803, 56.4.6.4.2.3 Status Phase, IMX6ULLRM
	if n == 0 && dir == IN {
		err = hw.transferDTD(n, OUT, true, []byte{})

		if err != nil {
			return
		}
	}

	return
}

func (hw *usb) ack(n int) (err error) {
	return hw.transferDTD(n, IN, true, nil)
}

func (hw *usb) stall(n int, dir int) {
	ctrl := (*uint32)(unsafe.Pointer(uintptr(hw.epctrl + uint32(4*n))))

	if dir == IN {
		reg.Set(ctrl, ENDPTCTRL_TXS)
	} else {
		reg.Set(ctrl, ENDPTCTRL_RXS)
	}
}

func (hw *usb) enable(n int, dir int, transferType int) {
	if n == 0 {
		// EP0 does not need enabling (p3790, IMX6ULLRM)
		return
	}

	log.Printf("imx6_usb: enabling EP%d.%d\n", n, dir)

	// TODO: clean specific cache lines instead
	cache.FlushData()

	ctrl := (*uint32)(unsafe.Pointer(uintptr(hw.epctrl + uint32(4*n))))
	c := *ctrl

	if dir == IN {
		reg.Set(&c, ENDPTCTRL_TXE)
		reg.Set(&c, ENDPTCTRL_TXR)
		reg.SetN(&c, ENDPTCTRL_TXT, 0b11, uint32(transferType))
		reg.Clear(&c, ENDPTCTRL_TXS)

		if reg.Get(ctrl, ENDPTCTRL_RXE, 0b1) == 0 {
			// see note at p3879 of IMX6ULLRM
			reg.SetN(&c, ENDPTCTRL_RXT, 0b11, BULK)
		}
	} else {
		reg.Set(&c, ENDPTCTRL_RXE)
		reg.Set(&c, ENDPTCTRL_RXR)
		reg.SetN(&c, ENDPTCTRL_RXT, 0b11, uint32(transferType))
		reg.Clear(&c, ENDPTCTRL_RXS)

		if reg.Get(ctrl, ENDPTCTRL_TXE, 0b1) == 0 {
			// see note at p3879 of IMX6ULLRM
			reg.SetN(&c, ENDPTCTRL_TXT, 0b11, BULK)
		}
	}

	*ctrl = c
}
