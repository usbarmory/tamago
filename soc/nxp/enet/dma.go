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
	BD_TX_ST_TC = 10 // Transmit CRC
)

// bufferDescriptor represents a legacy FEC receive/transmit buffer descriptor
// (p1012, 22.6.13 Legacy buffer descriptors, IMX6ULLRM).
type bufferDescriptor struct {
	Length uint16
	Status uint16
	Addr   uint32

	// DMA buffers
	desc []byte
	data []byte
}

func (bd *bufferDescriptor) Bytes() []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, bd.Length)
	binary.Write(buf, binary.LittleEndian, bd.Status)
	binary.Write(buf, binary.LittleEndian, bd.Addr)

	return buf.Bytes()
}

func (bd *bufferDescriptor) Data() (buf []byte) {
	buf = make([]byte, bd.Length-4)
	copy(buf, bd.data)
	return
}

type bufferDescriptorRing struct {
	bds   []*bufferDescriptor
	index int
	size  int
}

func (ring *bufferDescriptorRing) init(rx bool, n int) uint32 {
	ring.size = n
	ring.bds = make([]*bufferDescriptor, n)

	// To avoid excessive DMA region fragmentation, a single allocation
	// reserves all descriptors and data pointers which are slices for each
	// entry.

	descSize := len((&bufferDescriptor{}).Bytes())
	ptr, desc := dma.Reserve(n*descSize, bufferAlign)

	dataSize := MTU + (bufferAlign - (MTU % bufferAlign))
	addr, data := dma.Reserve(n*dataSize, bufferAlign)

	for i := 0; i < n; i++ {
		off := dataSize * i

		bd := &bufferDescriptor{
			Addr: uint32(addr) + uint32(off),
			data: data[off : off+dataSize],
		}

		if rx {
			bd.Status |= 1 << BD_RX_ST_E
		}

		if i == n-1 {
			bd.Status |= 1 << BD_ST_W
		}

		off = descSize * i
		bd.desc = desc[off : off+descSize]
		copy(bd.desc, bd.Bytes())

		ring.bds[i] = bd
	}

	return uint32(ptr)
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

func (ring *bufferDescriptorRing) pop() (data []byte) {
	bd := ring.bds[ring.index]

	bd.Length = uint16(bd.desc[0])
	bd.Length |= uint16(bd.desc[1]) << 8

	bd.Status = uint16(bd.desc[2])
	bd.Status |= uint16(bd.desc[3]) << 8

	if bd.Status&(1<<BD_RX_ST_E) != 0 {
		return
	}

	// set empty
	bd.desc[3] |= (1 << BD_RX_ST_E) >> 8

	ring.next()

	if bd.Length > MTU {
		print("enet: frame > MTU\n")
	}

	if bd.Status&(1<<BD_ST_L) == 0 {
		print("enet: frame not last\n")
	}

	if bd.Status&(1<<BD_RX_ST_CR) == 0 {
		return bd.Data()
	}

	return
}

func (ring *bufferDescriptorRing) push(data []byte) {
	bd := ring.bds[ring.index]

	if uint16(bd.desc[3]<<8)&(1<<BD_TX_ST_R) != 0 {
		print("enet: frame not sent\n")
	}

	bd.Length = uint16(len(data))
	bd.Status = (1 << BD_ST_L) | (1 << BD_TX_ST_TC)

	bd.desc[0] = byte(bd.Length & 0xff)
	bd.desc[1] = byte((bd.Length & 0xff00) >> 8)

	bd.desc[2] = byte((bd.Status & 0xff))
	bd.desc[3] = byte((bd.Status & 0xff00) >> 8)

	copy(bd.data, data)

	if ring.next() {
		bd.desc[3] |= (1 << BD_ST_W) >> 8
	}

	// set ready
	bd.desc[3] |= (1 << BD_TX_ST_R) >> 8
}

// Rx receives a single Ethernet frame, excluding the checksum, from the MAC
// controller ring buffer.
func (hw *ENET) Rx() (buf []byte) {
	hw.Lock()
	defer hw.Unlock()

	buf = hw.rx.pop()
	reg.Set(hw.rdar, RDAR_ACTIVE)

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

	hw.tx.push(buf)
	reg.Set(hw.tdar, TDAR_ACTIVE)
}
