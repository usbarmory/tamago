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
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"unsafe"

	"github.com/f-secure-foundry/tamago/imx6/internal/bits"
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
	next   uint32
	token  uint32
	buffer [5]uint32

	// DMA buffer pointers
	_dtd   uint32
	_pages uint32
}

// dQH implements
// p3784, 56.4.5.1 Endpoint Queue Head (dQH), IMX6ULLRM.
type dQH struct {
	info    uint32
	current uint32
	next    uint32
	token   uint32
	buffer  [5]uint32

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
	ep.List = (*[MAX_ENDPOINTS * 2]dQH)(ep.buf.Ptr())
}

// get returns the Endpoint Queue Head (dQH)
func (ep *EndPointList) get(n int, dir int) dQH {
	cache.FlushData()
	return ep.List[n*2+dir]
}

// next sets the next endpoint transfer pointer
func (ep *EndPointList) next(n int, dir int, next uint32) {
	ep.List[n*2+dir].next = next
}

// reset clears the endpoint status
func (ep *EndPointList) reset(n int, dir int) {
	cache.FlushData()
	bits.SetN(&ep.List[n*2+dir].token, 0, 0xff, 0)
}

// set configures a queue head as described in
// p3784, 56.4.5.1 Endpoint Queue Head, IMX6ULLRM.
func (ep *EndPointList) set(n int, dir int, max int, zlt int, mult int) {
	off := n*2 + dir

	// Mult
	bits.SetN(&ep.List[off].info, 30, 0b11, uint32(mult))
	// zlt
	bits.SetN(&ep.List[off].info, 29, 0b1, uint32(zlt))
	// Maximum Packet Length
	bits.SetN(&ep.List[off].info, 16, 0x7ff, uint32(max))

	if n == 0 && dir == IN {
		// interrupt on setup (ios)
		bits.Set(&ep.List[off].info, 15)
	}

	// Total bytes
	bits.SetN(&ep.List[off].token, 16, 0xffff, 0)
	// interrupt on completion (ioc)
	bits.Set(&ep.List[off].token, 15)
	// multiplier override (MultO)
	bits.SetN(&ep.List[off].token, 10, 0b11, 0)
}

// addDTD configures an endpoint transfer descriptor as described in
// p3787, 56.4.5.2 Endpoint Transfer Descriptor (dTD), IMX6ULLRM.
func buildDTD(n int, dir int, ioc bool, data []byte) (dtd *dTD) {
	// p3809, 56.4.6.6.2 Building a Transfer Descriptor, IMX6ULLRM
	dtd = &dTD{}

	// interrupt on completion (ioc)
	if ioc {
		bits.Set(&dtd.token, 15)
	} else {
		bits.Clear(&dtd.token, 15)
	}

	// invalidate next pointer
	dtd.next = 0x1
	// multiplier override (MultO)
	bits.SetN(&dtd.token, 10, 0b11, 0)
	// active status
	bits.Set(&dtd.token, 7)
	// total bytes
	bits.SetN(&dtd.token, 16, 0xffff, uint32(len(data)))

	dtd._pages = mem.Alloc(data, DTD_PAGE_SIZE)

	for n := 0; n < DTD_PAGES; n++ {
		dtd.buffer[n] = dtd._pages + uint32(DTD_PAGE_SIZE*n)
	}

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, dtd)

	// skip internal DMA buffer pointers
	dtd._dtd = mem.Alloc(buf.Bytes()[0:28], 32)

	return
}

// transferDTD manages a transfer using transfer descriptors (dTDs) as
// described in p3809, 56.4.6.6 Managing Transfers with Transfer Descriptors,
// IMX6ULLRM.
func (hw *usb) transferDTD(n int, dir int, ioc bool, in []byte) (out []byte, err error) {
	var data []byte
	var dtds []*dTD

	max := DTD_PAGES * DTD_PAGE_SIZE

	if dir == IN {
		data = in
	} else {
		data = make([]byte, max)
	}

	dtdLength := len(data)

	if dtdLength > max {
		dtdLength = max
	}

	dtd := buildDTD(n, dir, ioc, data[0:dtdLength])
	dtds = append(dtds, dtd)

	for i := dtdLength; i < len(data); i += dtdLength {
		size := i + dtdLength

		if size > len(data) {
			size = len(data)
		}

		next := buildDTD(n, dir, ioc, data[i:size])

		// treat dtd.next as a register within the dtd DMA buffer
		reg.Write(dtd._dtd, next._dtd)

		dtd = next
		dtds = append(dtds, next)
	}

	// set dQH head pointer
	hw.EP.next(n, dir, dtds[0]._dtd)
	// reset dQH status
	hw.EP.reset(n, dir)

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
		defer mem.Free(dtd._dtd)
		defer mem.Free(dtd._pages)

		// Treat dtd.token as a register within the dtd DMA buffer
		token := dtd._dtd + 4

		reg.Wait(token, 7, 0b1, 0)

		if status := reg.Get(token, 0, 0xff); status != 0x00 {
			return nil, fmt.Errorf("error status for dTD #%d, %x", i, status)
		}

		// p3787 "This field is decremented by the number of bytes
		// actually moved during the transaction", IMX6ULLRM.
		size := dtdLength - int(reg.Get(token, 16, 0xffff))

		if n != 0 && dir == OUT && size != 0 {
			out = append(out, mem.Read(dtd._pages)[0:size]...)
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
	ctrl := hw.epctrl + uint32(4*n)

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

	ctrl := hw.epctrl + uint32(4*n)
	c := reg.Read(ctrl)

	if dir == IN {
		c |= (1 << ENDPTCTRL_TXE)
		c |= (1 << ENDPTCTRL_TXR)
		c = (c & (^(uint32(0b11) << ENDPTCTRL_TXT))) | (uint32(transferType) << ENDPTCTRL_TXT)
		c &^= (1 << ENDPTCTRL_TXS)

		if reg.Get(ctrl, ENDPTCTRL_RXE, 0b1) == 0 {
			// see note at p3879 of IMX6ULLRM
			c = (c & (^(uint32(0b11) << ENDPTCTRL_RXT))) | (BULK << ENDPTCTRL_RXT)
		}
	} else {
		c |= (1 << ENDPTCTRL_RXE)
		c |= (1 << ENDPTCTRL_RXR)
		c = (c & (^(uint32(0b11) << ENDPTCTRL_RXT))) | (uint32(transferType) << ENDPTCTRL_RXT)
		c &^= (1 << ENDPTCTRL_RXS)

		if reg.Get(ctrl, ENDPTCTRL_TXE, 0b1) == 0 {
			// see note at p3879 of IMX6ULLRM
			c = (c & (^(uint32(0b11) << ENDPTCTRL_TXT))) | (BULK << ENDPTCTRL_TXT)
		}
	}

	reg.Write(ctrl, c)
}
