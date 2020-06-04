// NXP USBOH3USBO2 / USBPHY driver
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package usb

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"github.com/f-secure-foundry/tamago/dma"
	"github.com/f-secure-foundry/tamago/internal/bits"
	"github.com/f-secure-foundry/tamago/internal/reg"
)

// Endpoint constants
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

	// DMA pointer for dTD structure
	_dtd uint32
	// DMA pointer for dTD transfer buffer
	_buf uint32
	// transfer buffer size
	_size uint32
}

func (dtd *dTD) free() {
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
func (hw *USB) initEP() {
	var epList EndpointList
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, &epList)
	epListAddr := dma.Alloc(buf.Bytes(), DQH_LIST_ALIGN)

	// set endpoint queue head
	reg.Write(hw.eplist, epListAddr)
}

// setEP configures a queue head as described in
// p3784, 56.4.5.1 Endpoint Queue Head, IMX6ULLRM.
func (hw *USB) setEP(n int, dir int, max int, zlt bool, mult int) {
	dqh := dQH{}

	// Maximum Packet Length
	bits.SetN(&dqh.Info, 16, 0x7ff, uint32(max))

	if !zlt {
		// Zero Length Termination must be disabled for multi dTD
		// requests.
		bits.SetN(&dqh.Info, 29, 1, 1)
	}

	// Mult
	bits.SetN(&dqh.Info, 30, 0b11, uint32(mult))

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
func (hw *USB) getEP(n int, dir int) (dqh dQH) {
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
func (hw *USB) nextDTD(n int, dir int, next uint32) {
	offset := (n*2 + dir) * DQH_SIZE
	dqh := reg.Read(hw.eplist) + uint32(offset)

	// set next dTD
	reg.Write(dqh+uint32(DQH_NEXT), next)
	// reset endpoint status (active and halt bits)
	reg.SetN(dqh+uint32(DQH_TOKEN), 6, 0b11, 0b00)
}

// buildDTD configures an endpoint transfer descriptor as described in
// p3787, 56.4.5.2 Endpoint Transfer Descriptor (dTD), IMX6ULLRM.
func buildDTD(n int, dir int, ioc bool, addr uint32, size int) (dtd *dTD) {
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
	bits.SetN(&dtd.Token, 16, 0xffff, uint32(size))

	dtd._buf = addr
	dtd._size = uint32(size)

	for n := 0; n < DTD_PAGES; n++ {
		dtd.Buffer[n] = dtd._buf + uint32(DTD_PAGE_SIZE*n)
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
func (hw *USB) transferDTD(n int, dir int, ioc bool, buf []byte) (out []byte, err error) {
	var dtds []*dTD
	dtdLength := DTD_PAGES * DTD_PAGE_SIZE

	if dir == OUT && buf == nil {
		buf = make([]byte, dtdLength)
	}

	transferSize := len(buf)

	if transferSize < dtdLength {
		dtdLength = transferSize
	}

	pages := dma.Alloc(buf, DTD_PAGE_SIZE)
	defer dma.Free(pages)

	dtd := buildDTD(n, dir, ioc, pages, dtdLength)
	defer dtd.free()

	dtds = append(dtds, dtd)

	for i := dtdLength; i < transferSize; i += dtdLength {
		size := dtdLength

		if i+size > transferSize {
			size = transferSize - i
		}

		next := buildDTD(n, dir, ioc, pages+uint32(i), size)
		defer next.free()

		// treat dtd.next as a register within the dtd DMA buffer
		reg.Write(dtd._dtd+DTD_NEXT, next._dtd)

		dtd = next
		dtds = append(dtds, next)
	}

	// set dQH head pointer
	hw.nextDTD(n, dir, dtds[0]._dtd)

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

	size := 0

	for i, dtd := range dtds {
		// treat dtd.token as a register within the dtd DMA buffer
		token := dtd._dtd + DTD_TOKEN

		// The hardware might delay status update after completion,
		// therefore best to wait for the active bit (7) to clear.
		inactive := reg.WaitFor(5*time.Second, token, 7, 1, 0)
		dtdToken := reg.Read(token)

		if !inactive {
			return nil, fmt.Errorf("dTD[%d] timeout waiting for completion, token:%#x", i, dtdToken)
		}

		if (dtdToken & 0xff) != 0 {
			return nil, fmt.Errorf("dTD[%d] error status, token:%#x", i, dtdToken)
		}

		// p3787 "This field is decremented by the number of bytes
		// actually moved during the transaction", IMX6ULLRM.
		rest := dtdToken >> 16
		n := int(dtd._size - rest)

		if dir == IN && rest > 0 {
			return nil, fmt.Errorf("dTD[%d] partial transfer (%d/%d bytes)", i, n, dtd._size)
		}

		size += n
	}

	if n != 0 && dir == OUT {
		dma.Read(pages, 0, buf)
		out = buf[0:size]
	}

	return
}

func (hw *USB) tx(n int, ioc bool, in []byte) (err error) {
	_, err = hw.transferDTD(n, IN, ioc, in)

	if err != nil {
		return
	}

	// p3803, 56.4.6.4.2.3 Status Phase, IMX6ULLRM
	if n == 0 {
		_, err = hw.transferDTD(n, OUT, false, nil)
	}

	return
}

func (hw *USB) rx(n int, ioc bool, buf []byte) (out []byte, err error) {
	return hw.transferDTD(n, OUT, ioc, buf)
}

func (hw *USB) ack(n int) (err error) {
	_, err = hw.transferDTD(n, IN, false, nil)
	return
}

func (hw *USB) stall(n int, dir int) {
	ctrl := hw.epctrl + uint32(4*n)

	if dir == IN {
		reg.Set(ctrl, ENDPTCTRL_TXS)
	} else {
		reg.Set(ctrl, ENDPTCTRL_RXS)
	}
}

func (hw *USB) enable(n int, dir int, transferType int) {
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
