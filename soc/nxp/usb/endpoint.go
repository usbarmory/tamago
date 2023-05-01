// NXP USBOH3USBO2 / USBPHY driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package usb

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/internal/reg"
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

	DQH_INFO  = 0
	INFO_MULT = 30
	INFO_ZLT  = 29
	INFO_MPL  = 16
	INFO_IOS  = 15

	DQH_CURRENT = 4
	DQH_NEXT    = 8
	DQH_TOKEN   = 12

	// p3787, 56.4.5.2 Endpoint Transfer Descriptor (dTD), IMX6ULLRM
	DTD_ALIGN     = 32
	DTD_SIZE      = 28
	DTD_PAGES     = 5
	DTD_PAGE_SIZE = 4096
	DTD_NEXT      = 0

	DTD_TOKEN    = 4
	TOKEN_TOTAL  = 16
	TOKEN_IOC    = 15
	TOKEN_MULTO  = 10
	TOKEN_ACTIVE = 7
	TOKEN_STATUS = 0
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

// endpointList implements
// p3783, 56.4.5 Device Data Structures, IMX6ULLRM.
type endpointList [MAX_ENDPOINTS * 2]dQH

// initQH initializes the endpoint queue head list
func (hw *USB) initQH() {
	var epList endpointList
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, &epList)
	hw.epListAddr = uint32(dma.Alloc(buf.Bytes(), DQH_LIST_ALIGN))

	// set endpoint queue head
	reg.Write(hw.eplist, hw.epListAddr)
}

// set configures an endpoint queue head as described in
// p3784, 56.4.5.1 Endpoint Queue Head, IMX6ULLRM.
func (hw *USB) set(n int, dir int, max int, zlt bool, mult int) {
	dqh := dQH{}

	// Mult
	bits.SetN(&dqh.Info, INFO_MULT, 0b11, uint32(mult))
	// Maximum Packet Length
	bits.SetN(&dqh.Info, INFO_MPL, 0x7ff, uint32(max))

	if !zlt {
		// Zero Length Termination must be disabled for multi dTD
		// requests.
		bits.SetN(&dqh.Info, INFO_ZLT, 1, 1)
	}

	if n == 0 {
		// interrupt on setup (ios)
		bits.Set(&dqh.Info, INFO_IOS)
	}

	// Total bytes
	bits.SetN(&dqh.Token, TOKEN_TOTAL, 0xffff, 0)
	// interrupt on completion (ioc)
	bits.Set(&dqh.Token, TOKEN_IOC)
	// multiplier override (MultO)
	bits.SetN(&dqh.Token, TOKEN_MULTO, 0b11, 0)

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, &dqh)

	offset := (n*2 + dir) * DQH_SIZE
	dma.Write(uint(hw.epListAddr), offset, buf.Bytes())

	hw.dQH[n][dir] = hw.epListAddr + uint32(offset)
}

// enable enables an endpoint.
func (hw *USB) enable(n int, dir int, transferType int) {
	if n == 0 {
		// EP0 does not need enabling (p3790, IMX6ULLRM)
		return
	}

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

// clear resets the endpoint status (active and halt bits)
func (hw *USB) clear(n int, dir int) {
	token := hw.dQH[n][dir] + DQH_TOKEN
	reg.SetN(token, TOKEN_STATUS, 0xc0, 0)
}

// qh returns the Endpoint Queue Head (dQH)
func (hw *USB) qh(n int, dir int) (dqh dQH) {
	buf := make([]byte, DQH_SIZE)
	dma.Read(uint(hw.dQH[n][dir]), 0, buf)

	err := binary.Read(bytes.NewReader(buf), binary.LittleEndian, &dqh)

	if err != nil {
		panic(err)
	}

	return
}

// nextDTD sets the next endpoint transfer pointer
func (hw *USB) nextDTD(n int, dir int, dtd uint32) {
	dqh := hw.dQH[n][dir]
	next := dqh + DQH_NEXT

	// wait for endpoint status to be cleared
	reg.Wait(dqh+DQH_TOKEN, TOKEN_STATUS, 0xc0, 0)
	// set next dTD
	reg.Write(next, dtd)
}

// buildDTD configures an endpoint transfer descriptor as described in
// p3787, 56.4.5.2 Endpoint Transfer Descriptor (dTD), IMX6ULLRM.
func buildDTD(n int, dir int, addr uint32, size int) (dtd *dTD) {
	// p3809, 56.4.6.6.2 Building a Transfer Descriptor, IMX6ULLRM
	dtd = &dTD{}

	// interrupt on completion (ioc)
	bits.Set(&dtd.Token, TOKEN_IOC)

	// invalidate next pointer
	dtd.Next = 1
	// multiplier override (MultO)
	bits.SetN(&dtd.Token, TOKEN_MULTO, 0b11, 0)
	// active status
	bits.Set(&dtd.Token, TOKEN_ACTIVE)
	// total bytes
	bits.SetN(&dtd.Token, TOKEN_TOTAL, 0xffff, uint32(size))

	dtd._buf = addr
	dtd._size = uint32(size)

	for n := 0; n < DTD_PAGES; n++ {
		dtd.Buffer[n] = dtd._buf + DTD_PAGE_SIZE*uint32(n)
	}

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, dtd)

	// skip internal DMA buffer pointers
	dtd._dtd = uint32(dma.Alloc(buf.Bytes()[0:DTD_SIZE], DTD_ALIGN))

	return
}

// checkDTD verifies transfer descriptor completion as describe in
// p3800, 56.4.6.4.1 Interrupt/Bulk Endpoint Operational Model, IMX6ULLRM
// p3811, 56.4.6.6.4 Transfer Completion, IMX6ULLRM.
func (hw *USB) checkDTD(n int, dir int, dtds []*dTD) (size int, err error) {
	for i, dtd := range dtds {
		// treat dtd.token as a register within the dtd DMA buffer
		token := dtd._dtd + DTD_TOKEN

		// wait for active bit to be cleared
		reg.WaitSignal(hw.exit, token, TOKEN_ACTIVE, 1, 0)

		dtdToken := reg.Read(token)

		if (dtdToken & 0xff) != 0 {
			return 0, fmt.Errorf("dTD[%d] error status, token:%#x", i, dtdToken)
		}

		// p3787 "This field is decremented by the number of bytes
		// actually moved during the transaction", IMX6ULLRM.
		rest := dtdToken >> TOKEN_TOTAL
		n := int(dtd._size - rest)

		if dir == IN && rest > 0 {
			return 0, fmt.Errorf("dTD[%d] partial transfer (%d/%d bytes)", i, n, dtd._size)
		}

		size += n
	}

	return
}

// transfer initates a transfer using transfer descriptors (dTDs) as described in
// p3810, 56.4.6.6.3 Executing A Transfer Descriptor, IMX6ULLRM.
func (hw *USB) transfer(n int, dir int, buf []byte) (out []byte, err error) {
	var dtds []*dTD
	var prev *dTD
	var i int

	// hw.prime IN:ENDPTPRIME_PETB+n    OUT:ENDPTPRIME_PERB+n
	// hw.pos   IN:ENDPTCOMPLETE_ETCE+n OUT:ENDPTCOMPLETE_ERCE+n
	pos := (dir * 16) + n

	dtdLength := DTD_PAGES * DTD_PAGE_SIZE

	if dir == OUT && buf == nil {
		buf = make([]byte, dtdLength)
	}

	transferSize := len(buf)

	pages := dma.Alloc(buf, DTD_PAGE_SIZE)
	defer dma.Free(pages)

	// loop condition to account for zero transferSize
	for add := true; add; add = i < transferSize {
		prime := false
		size := dtdLength

		if i+size > transferSize {
			size = transferSize - i
		}

		dtd := buildDTD(n, dir, uint32(pages)+uint32(i), size)
		defer dma.Free(uint(dtd._dtd))

		if i == 0 {
			prime = true
		} else {
			// treat dtd.next as a register within the dtd DMA buffer
			reg.Write(prev._dtd+DTD_NEXT, dtd._dtd)
			prime = reg.Get(hw.prime, pos, 1) == 0 && reg.Get(hw.stat, pos, 1) == 0
		}

		if prime {
			// reset endpoint status
			hw.clear(n, dir)
			// set dQH head pointer
			hw.nextDTD(n, dir, dtd._dtd)
			// prime endpoint
			reg.Set(hw.prime, pos)
		}

		prev = dtd
		dtds = append(dtds, dtd)

		i += dtdLength
	}

	// wait for priming completion
	reg.Wait(hw.prime, pos, 1, 0)

	if hw.event != nil && n != 0 {
		// wait for completion (event)
		for reg.Get(hw.complete, pos, 1) != 1 {
			hw.event.L.Lock()
			hw.event.Wait()
			hw.event.L.Unlock()
		}
	} else {
		// wait for completion (poll)
		reg.WaitSignal(hw.exit, hw.complete, pos, 1, 1)
	}

	// clear completion
	reg.Write(hw.complete, 1<<pos)

	size, err := hw.checkDTD(n, dir, dtds)

	if n != 0 && dir == OUT && buf != nil {
		out = buf[0:size]
		dma.Read(pages, 0, out)
	}

	return
}

// ack transmits a zero length packet to the host through an IN endpoint
func (hw *USB) ack(n int) (err error) {
	_, err = hw.transfer(n, IN, nil)
	return
}

// tx transmits a data buffer to the host through an IN endpoint
func (hw *USB) tx(n int, in []byte) (err error) {
	_, err = hw.transfer(n, IN, in)

	// p3803, 56.4.6.4.2.3 Status Phase, IMX6ULLRM
	if err == nil && n == 0 {
		_, err = hw.transfer(n, OUT, nil)
	}

	return
}

// rx receives a data buffer from the host through an OUT endpoint
func (hw *USB) rx(n int, buf []byte) (out []byte, err error) {
	return hw.transfer(n, OUT, buf)
}

// stall forces the endpoint to return a STALL handshake to the host
func (hw *USB) stall(n int, dir int) {
	ctrl := hw.epctrl + uint32(4*n)

	if dir == IN {
		reg.Set(ctrl, ENDPTCTRL_TXS)
	} else {
		reg.Set(ctrl, ENDPTCTRL_RXS)
	}
}

// reset forces data PID synchronization between host and device
func (hw *USB) reset(n int, dir int) {
	if n == 0 {
		// EP0 does not have data toggle reset
		return
	}

	ctrl := hw.epctrl + uint32(4*n)

	if dir == IN {
		reg.Set(ctrl, ENDPTCTRL_TXR)
	} else {
		reg.Set(ctrl, ENDPTCTRL_RXR)
	}
}
