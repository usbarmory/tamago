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
	"time"

	"github.com/f-secure-foundry/tamago/internal/bits"
	"github.com/f-secure-foundry/tamago/internal/dma"
	"github.com/f-secure-foundry/tamago/internal/reg"
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

	// p3784, 56.4.5.1 Endpoint Queue Head (dQH), IMX6ULLRM
	DQH_LIST_ALIGN = 2048
	DQH_ALIGN      = 64
	DQH_SIZE       = 64
	DQH_INFO       = 0
	DQH_CURRENT    = 4
	DQH_NEXT       = 8
	DQH_TOKEN      = 12

	// p3787, 56.4.5.2 Endpoint Transfer Descriptor (dTD), IMX6ULLRM
	DTD_ALIGN     = 32
	DTD_SIZE      = 28
	DTD_PAGES     = 5
	DTD_PAGE_SIZE = 4096
	DTD_NEXT      = 0
	DTD_TOKEN     = 4
)

// dTD implements
// p3787, 56.4.5.2 Endpoint Transfer Descriptor (dTD), IMX6ULLRM.
type dTD struct {
	Next   uint32
	Token  uint32
	Buffer [5]uint32

	// DMA buffer pointers
	_dtd   uint32
	_pages uint32
}

func (dtd *dTD) free() {
	dma.Free(dtd._pages)
	dma.Free(dtd._dtd)
}

// dQH implements
// p3784, 56.4.5.1 Endpoint Queue Head (dQH), IMX6ULLRM.
type dQH struct {
	Info    uint32
	Current uint32
	Next    uint32
	Token   uint32
	Buffer  [5]uint32

	// reserved
	_ uint32

	// The Set-up Buffer will be filled by hardware, note that after this
	// happens endianess needs to be adjusted with SetupData.swap().
	Setup SetupData

	// We align only the first queue entry, so we need a 4*uint32 gap to
	// maintain 64-byte boundaries.
	_ [4]uint32
}

// EndpointList implements
// p3783, 56.4.5 Device Data Structures, IMX6ULLRM.
type EndpointList [MAX_ENDPOINTS * 2]dQH

// initEP initializes the endpoint queue head list
func (hw *usb) initEP() {
	var epList EndpointList
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, &epList)
	epListAddr := dma.Alloc(buf.Bytes(), DQH_LIST_ALIGN)

	// set endpoint queue head
	reg.Write(hw.eplist, epListAddr)
}

// setEP configures a queue head as described in
// p3784, 56.4.5.1 Endpoint Queue Head, IMX6ULLRM.
func (hw *usb) setEP(n int, dir int, max int, mult int) {
	dqh := dQH{}

	// Mult
	bits.SetN(&dqh.Info, 30, 0b11, uint32(mult))
	// zlt
	bits.SetN(&dqh.Info, 29, 1, 0)
	// Maximum Packet Length
	bits.SetN(&dqh.Info, 16, 0x7ff, uint32(max))

	if n == 0 && dir == IN {
		// interrupt on setup (ios)
		bits.Set(&dqh.Info, 15)
	}

	// Total bytes
	bits.SetN(&dqh.Token, 16, 0xffff, 0)
	// interrupt on completion (ioc)
	bits.Set(&dqh.Token, 15)
	// multiplier override (MultO)
	bits.SetN(&dqh.Token, 10, 0b11, 0)

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, &dqh)

	epListAddr := reg.Read(hw.eplist)
	offset := (n*2 + dir) * DQH_SIZE
	dma.Write(epListAddr, buf.Bytes(), offset)
}

// getEP returns an Endpoint Queue Head (dQH)
func (hw *usb) getEP(n int, dir int) (dqh dQH) {
	epListAddr := reg.Read(hw.eplist)
	offset := (n*2 + dir) * DQH_SIZE

	buf := make([]byte, DQH_SIZE)
	dma.Read(epListAddr, offset, buf)

	err := binary.Read(bytes.NewReader(buf), binary.LittleEndian, &dqh)

	if err != nil {
		panic(err)
	}

	return
}

// next sets the next endpoint transfer pointer
func (hw *usb) nextDTD(n int, dir int, next uint32, zlt int) {
	offset := (n*2 + dir) * DQH_SIZE
	dqh := reg.Read(hw.eplist) + uint32(offset)

	// set next dTD
	reg.Write(dqh+uint32(DQH_NEXT), next)
	// zlt
	reg.SetN(dqh+uint32(DQH_INFO), 29, 1, uint32(zlt))
	// reset endpoint status (active and halt bits)
	reg.SetN(dqh+uint32(DQH_TOKEN), 6, 0b11, 0b00)
}

// addDTD configures an endpoint transfer descriptor as described in
// p3787, 56.4.5.2 Endpoint Transfer Descriptor (dTD), IMX6ULLRM.
func buildDTD(n int, dir int, ioc bool, data []byte) (dtd *dTD) {
	// p3809, 56.4.6.6.2 Building a Transfer Descriptor, IMX6ULLRM
	dtd = &dTD{}

	// interrupt on completion (ioc)
	if ioc {
		bits.Set(&dtd.Token, 15)
	} else {
		bits.Clear(&dtd.Token, 15)
	}

	// invalidate next pointer
	dtd.Next = 0b1
	// multiplier override (MultO)
	bits.SetN(&dtd.Token, 10, 0b11, 0)
	// active status
	bits.Set(&dtd.Token, 7)
	// total bytes
	bits.SetN(&dtd.Token, 16, 0xffff, uint32(len(data)))

	dtd._pages = dma.Alloc(data, DTD_PAGE_SIZE)

	for n := 0; n < DTD_PAGES; n++ {
		dtd.Buffer[n] = dtd._pages + uint32(DTD_PAGE_SIZE*n)
	}

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, dtd)

	// skip internal DMA buffer pointers
	dtd._dtd = dma.Alloc(buf.Bytes()[0:DTD_SIZE], DTD_ALIGN)

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
	defer dtd.free()

	dtds = append(dtds, dtd)

	for i := dtdLength; i < len(data); i += dtdLength {
		size := i + dtdLength

		if size > len(data) {
			size = len(data)
		}

		next := buildDTD(n, dir, ioc, data[i:size])
		defer next.free()

		// treat dtd.next as a register within the dtd DMA buffer
		reg.Write(dtd._dtd+DTD_NEXT, next._dtd)

		dtd = next
		dtds = append(dtds, next)
	}

	zlt := 0

	// If more than one dTD is sent we see that the automatic zlt is
	// causing issues, for now we leave up to the application sending zlt
	// on large transfers.
	if dir == IN && len(dtds) > 1 {
		zlt = 1
	}

	// set dQH head pointer
	hw.nextDTD(n, dir, dtds[0]._dtd, zlt)

	// hw.prime IN:ENDPTPRIME_PETB+n    OUT:ENDPTPRIME_PERB+n
	// hw.pos   IN:ENDPTCOMPLETE_ETCE+n OUT:ENDPTCOMPLETE_ERCE+n
	pos := (dir * 16) + n

	// prime endpoint
	reg.Set(hw.prime, pos)
	// wait for priming completion
	reg.Wait(hw.prime, pos, 1, 0)

	// wait for completion
	reg.Wait(hw.complete, pos, 1, 1)
	// clear completion
	reg.Write(hw.complete, 1<<pos)

	for i, dtd := range dtds {
		// treat dtd.token as a register within the dtd DMA buffer
		token := dtd._dtd + DTD_TOKEN

		// The hardware might delay status update after completion,
		// therefore best to wait for the active bit (7) to clear.
		inactive := reg.WaitFor(5*time.Second, token, 7, 1, 0)
		dtdToken := reg.Read(token)

		if !inactive {
			return nil, fmt.Errorf("dTD[%d] timeout waiting for completion (token:%x)", i, dtdToken)
		}

		if (dtdToken & 0xff) != 0 {
			return nil, fmt.Errorf("dTD[%d] error status (token:%x)", i, dtdToken)
		}

		// p3787 "This field is decremented by the number of bytes
		// actually moved during the transaction", IMX6ULLRM.
		size := dtdLength - int(dtdToken>>16)

		if n != 0 && dir == OUT && size != 0 {
			buf := make([]byte, size)
			dma.Read(dtd._pages, 0, buf)
			out = append(out, buf...)
		}

		if dir == IN && size != dtdLength {
			return nil, fmt.Errorf("dTD[%d] partial transfer (%d/%d bytes)", i, size, dtdLength)
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
		bits.Set(&c, ENDPTCTRL_TXE)
		bits.Set(&c, ENDPTCTRL_TXR)
		bits.SetN(&c, ENDPTCTRL_TXT, 0b11, uint32(transferType))
		bits.Clear(&c, ENDPTCTRL_TXS)

		if reg.Get(ctrl, ENDPTCTRL_RXE, 1) == 0 {
			// see note at p3879 of IMX6ULLRM
			bits.SetN(&c, ENDPTCTRL_RXT, 0b11, BULK)
		}
	} else {
		bits.Set(&c, ENDPTCTRL_RXE)
		bits.Set(&c, ENDPTCTRL_RXR)
		bits.SetN(&c, ENDPTCTRL_RXT, 0b11, uint32(transferType))
		bits.Clear(&c, ENDPTCTRL_RXS)

		if reg.Get(ctrl, ENDPTCTRL_TXE, 1) == 0 {
			// see note at p3879 of IMX6ULLRM
			bits.SetN(&c, ENDPTCTRL_TXT, 0b11, BULK)
		}
	}

	reg.Write(ctrl, c)
}
