// NXP USBOH3USBO2 / USBPHY driver
// https://github.com/f-secure-foundry/tamago
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

	"github.com/f-secure-foundry/tamago/imx6/internal/cache"
	"github.com/f-secure-foundry/tamago/imx6/internal/mem"
	"github.com/f-secure-foundry/tamago/imx6/internal/reg"
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
	// maintain 64-byte boundaries.
	_align [4]uint32
}

// EndPointList implements
// p3783, 56.4.5 Device Data Structures, IMX6ULLRM.
type EndPointList struct {
	List *[MAX_ENDPOINTS * 2]dQH

	buf *mem.AlignmentBuffer
}

func (ep *EndPointList) init() {
	ep.buf = mem.NewAlignmentBuffer(unsafe.Sizeof(ep.List), 2048)
	ep.List = (*[MAX_ENDPOINTS * 2]dQH)(unsafe.Pointer(ep.buf.Addr))
}

// get returns the Endpoint Queue Head (dQH)
func (ep *EndPointList) get(n int, dir int) dQH {
	// TODO: clean specific cache lines instead
	cache.FlushData()
	return ep.List[n*2+dir]
}

// next sets the next endpoint transfer pointer
func (ep *EndPointList) next(n int, dir int, next *dTD) {
	ep.List[n*2+dir].next = next
}

// max returns the endpoint Maximum Packet Length
func (ep *EndPointList) max(n int, dir int) int {
	return int(ep.List[n*2+dir].info>>16) & 0x7ff
}

// reset clears the endpoint status
func (ep *EndPointList) reset(n int, dir int) {
	reg.SetN(&ep.List[n*2+dir].token, 0, 0xff, 0)
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
func buildDTD(n int, dir int, ioc bool, data []byte) (dtd *dTD, dtdBuf *mem.AlignmentBuffer, pages *mem.AlignmentBuffer, err error) {
	size := len(data)

	if size > DTD_PAGES*DTD_PAGE_SIZE {
		return nil, nil, nil, errors.New("unsupported transfer size")
	}

	dtdBuf = mem.NewAlignmentBuffer(unsafe.Sizeof(dTD{}), 32)
	dtd = (*dTD)(unsafe.Pointer(dtdBuf.Addr))

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

	pages = mem.NewAlignmentBuffer(DTD_PAGE_SIZE*DTD_PAGES, DTD_PAGE_SIZE)
	mem.Copy(pages, data)

	// total bytes
	reg.SetN(&dtd.token, 16, 0xffff, uint32(size))

	for n := 0; n < DTD_PAGES; n++ {
		dtd.buffer[n] = pages.Addr + uintptr(DTD_PAGE_SIZE*n)
	}

	// invalidate next pointer
	dtd.next = (*dTD)(unsafe.Pointer(uintptr(1)))

	return
}

// transferDTD manages a transfer using transfer descriptors (dTDs) as
// described in p3809, 56.4.6.6 Managing Transfers with Transfer Descriptors,
// IMX6ULLRM.
func (hw *usb) transferDTD(n int, dir int, ioc bool, in []byte) (out []byte, err error) {
	var data []byte
	var dtds []*dTD

	// All aligned buffers must be kept around to avoid GC wiping them out,
	// it doesn't work simply including them in the dTD structure as
	// pointers.
	var dtdBufs []*mem.AlignmentBuffer
	var pktBufs []*mem.AlignmentBuffer

	max := hw.EP.max(n, dir)

	if dir == IN {
		data = in
	} else {
		data = make([]byte, max)
	}

	dtdLength := len(data)

	if n != 0 {
		// On non-control IN endpoints, for simplicity, we configure
		// each dTD exactly one packet (dTD.TotalBytes == dQH.MaxPacketLength).

		if dtdLength > max {
			dtdLength = max
		}
	}

	dtd, buf, pages, err := buildDTD(n, dir, ioc, data[0:dtdLength])

	if err != nil {
		return
	}

	// set dQH head pointer
	hw.EP.next(n, dir, dtd)
	// reset dQH status
	hw.EP.reset(n, dir)

	dtds = append(dtds, dtd)
	dtdBufs = append(dtdBufs, buf)
	pktBufs = append(pktBufs, pages)

	for i := dtdLength; i < len(data); i += dtdLength {
		size := i + dtdLength

		if size > len(data) {
			size = len(data)
		}

		next, buf, pages, err := buildDTD(n, dir, ioc, data[i:size])

		if err != nil {
			return nil, err
		}

		dtd.next = next
		dtd = next

		dtds = append(dtds, next)
		dtdBufs = append(dtdBufs, buf)
		pktBufs = append(pktBufs, pages)
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

	for i, dtd := range dtds {
		reg.Wait(&dtd.token, 7, 0b1, 0)

		if status := (dtd.token & 0xff); status != 0x00 {
			return nil, fmt.Errorf("error status for dTD #%d, %x", i, status)
		}

		// p3787 "This field is decremented by the number of bytes
		// actually moved during the transaction", IMX6ULLRM.
		size := dtdLength - int(reg.Get(&dtd.token, 16, 0xffff))

		if n != 0 && dir == OUT && size != 0 {
			out = append(out, pktBufs[i].Data()[0:size]...)
		}

		if dir == IN && size != dtdLength {
			return nil, fmt.Errorf("error status for dTD #%d, partial transfer (%d/%d bytes)", i, size, dtdLength)
		}
	}

	return
}

func (hw *usb) tx(n int, ioc bool, in []byte) (err error) {
	_, err = hw.transferDTD(n, IN, ioc, in)

	if err != nil {
		return
	}

	// p3803, 56.4.6.4.2.3 Status Phase, IMX6ULLRM
	if n == 0 {
		_, err = hw.transferDTD(n, OUT, true, nil)
	}

	return
}

func (hw *usb) rx(n int, ioc bool) (out []byte, err error) {
	return hw.transferDTD(n, OUT, ioc, nil)
}

func (hw *usb) ack(n int) (err error) {
	_, err = hw.transferDTD(n, IN, true, nil)
	return
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
