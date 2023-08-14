// NXP 10/100-Mbps Ethernet MAC (ENET)
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package enet

import (
	"bytes"
	"encoding/binary"

	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/internal/reg"
)

const (
	MTU             = 1518
	defaultRingSize = 16
	bufferAlign     = 64
)

// Common buffer descriptor fields
const (
	BD_ST_W = 13 // Wrap
	BD_ST_L = 11 // Last
)

// p1014, Table 22-35. Receive buffer descriptor field definitions, IMX6ULLRM
const (
	BD_RX_ST_E  = 15 // Empty
	BD_RX_ST_CR = 2  // Receive CRC or frame error
)

// p1017, Table 22-37. Enhanced transmit buffer descriptor field definitions, IMX6ULLRM
const (
	BD_TX_ST_R  = 15 // Ready
	BD_TX_ST_W  = 13 // Wrap
	BD_TX_ST_TC = 10 // Transmit CRC
)

// bufferDescriptor represents a legacy FEC receive/transmit buffer descriptor
// (p1012, 22.6.13 Legacy buffer descriptors, IMX6ULLRM).
type bufferDescriptor struct {
	Length uint16
	Status uint16
	Addr   uint32

	// DMA buffer
	buf []byte
}

func (bd *bufferDescriptor) bytes() []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, bd.Length)
	binary.Write(buf, binary.LittleEndian, bd.Status)
	binary.Write(buf, binary.LittleEndian, bd.Addr)

	return buf.Bytes()
}

func (bd *bufferDescriptor) data() (buf []byte) {
	buf = make([]byte, bd.Length-4)
	copy(buf, bd.buf)
	return
}

type bufferDescriptorRing struct {
	bds   []bufferDescriptor
	index int
	size  int

	// DMA buffer
	buf  []byte
	addr uint
}

func (ring *bufferDescriptorRing) init(rx bool, size int) (ptr uint32) {
	ring.size = size
	ring.bds = make([]bufferDescriptor, size)

	n := MTU + (bufferAlign - (MTU % bufferAlign))
	addr, buf := dma.Reserve(size*n, bufferAlign)

	for i := 0; i < len(ring.bds); i++ {
		off := n * i

		if rx {
			ring.bds[i].Status |= 1 << BD_RX_ST_E
		}

		ring.bds[i].Addr = uint32(addr) + uint32(off)
		ring.bds[i].buf = buf[off:off+n]
	}

	ring.bds[len(ring.bds)-1].Status |= 1 << BD_ST_W

	ring.addr, ring.buf = dma.Reserve(len(ring.bds)*8, bufferAlign)

	for i, bd := range ring.bds {
		copy(ring.buf[i*8:], bd.bytes())
	}

	return uint32(ring.addr)
}

func (ring *bufferDescriptorRing) next() (wrap bool) {
	wrap = ring.index == (ring.size - 1)

	if wrap {
		ring.index = 0
	} else {
		ring.index += 1
	}

	return
}

func (ring *bufferDescriptorRing) pop() (bd bufferDescriptor, data []byte) {
	off := ring.index * 8
	bd = ring.bds[ring.index]

	bd.Length = uint16(ring.buf[off+0])
	bd.Length |= uint16(ring.buf[off+1]) << 8

	bd.Status = uint16(ring.buf[off+2])
	bd.Status |= uint16(ring.buf[off+3]) << 8

	if bd.Status&(1<<BD_RX_ST_E) != 0 {
		return
	}

	data = bd.data()

	// set empty
	ring.buf[off+3] |= (1 << BD_RX_ST_E) >> 8

	ring.next()

	return
}

func (ring *bufferDescriptorRing) push(bd bufferDescriptor) {
	off := ring.index * 8

	ring.buf[off+0] = byte((len(bd.buf) & 0xff))
	ring.buf[off+1] = byte((len(bd.buf) & 0xff00) >> 8)

	ring.buf[off+2] = byte((bd.Status & 0xff))
	ring.buf[off+3] = byte((bd.Status & 0xff00) >> 8)

	copy(ring.bds[ring.index].buf, bd.buf)

	if ring.next() {
		ring.buf[off+3] |= (1 << BD_ST_W) >> 8
	}

	// set ready
	ring.buf[off+3] |= (1 << BD_TX_ST_R) >> 8
}

// Rx receives a single Ethernet frame, excluding the checksum, from the MAC
// controller ring buffer.
func (hw *ENET) Rx() (buf []byte) {
	hw.Lock()
	defer hw.Unlock()

	bd, data := hw.rx.pop()

	if bd.Status&(1<<BD_RX_ST_E) != 0 {
		return
	}

	reg.Set(hw.rdar, RDAR_ACTIVE)

	if bd.Length > MTU {
		print("enet: frame > MTU\n")
		return
	}

	if bd.Status&(1<<BD_ST_L) == 0 {
		print("enet: frame not last\n")
		return
	}

	if bd.Status&(1<<BD_RX_ST_CR) == 0 {
		buf = data
	}

	return
}

// Tx transmits a single Ethernet frame, the checksum is appended
// automatically and must not be included.
func (hw *ENET) Tx(buf []byte) {
	hw.Lock()
	defer hw.Unlock()

	if len(buf) > MTU {
		return
	}

	bd := bufferDescriptor{
		Length: uint16(len(buf)),
		Status: (1 << BD_ST_L) | (1 << BD_TX_ST_TC),
		buf:    buf,
	}

	hw.tx.push(bd)

	reg.Set(hw.tdar, TDAR_ACTIVE)
}
