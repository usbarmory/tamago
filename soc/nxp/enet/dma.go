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
	MTU               = 1518
	minFrameSizeBytes = 42
	defaultRingSize   = 16
	bufferAlign       = 64
)

// Common buffer descriptor fields
const (
	BD_ST_W = 13 // Wrap
	BD_ST_L = 11 // Last
)

// p1014, Table 22-35. Receive buffer descriptor field definitions, IMX6ULLRM
const (
	BD_RX_ST_E  = 15 // Empty
	BD_RX_ST_LG = 5  // Frame length violation
	BD_RX_ST_NO = 4  // Non-octet aligned frame
	BD_RX_ST_CR = 2  // CRC or frame error
	BD_RX_ST_OV = 1  // Overrun
	BD_RX_ST_TR = 0  // Frame truncated

	frameErrorMask = 1<<BD_RX_ST_CR | 1<<BD_RX_ST_LG | 1<<BD_RX_ST_NO | 1<<BD_RX_ST_OV | 1<<BD_RX_ST_TR
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

	stats *Stats

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

func (bd *bufferDescriptor) Valid() bool {
	valid := true

	if bd.Status&(1<<BD_ST_L) != 0 {
		if bd.Status&frameErrorMask != 0 {
			// Something wrong with the received frame, drop it.
			valid = false
			bd.stats.MACErrors.Add(1)
		}
	} else {
		// This is probably part of a jumbo frame or some other weirdness from the MAC.
		// We can't handle it, so drop it.
		valid = false
	}

	if bd.Length > MTU {
		// If bd.Length is larger than MTU, this probably means that we've been receiving
		// buffers with BD_ST_L set to zero indicating that they are chunks of a frame larger
		// than our ring-buffer data buffers can hold, and we're now processing the "final" buffer
		// of this frame.
		// In this case, bd.Length represents the length of the *whole frame* rather than the
		// length of data in this buffer, so we MUST NOT call db.Data below, as this situation is
		// not currently handled.
		valid = false
		bd.stats.FramesTooLarge.Add(1)
	} else if bd.Length < minFrameSizeBytes {
		valid = false
		bd.stats.FramesTooSmall.Add(1)
	}

	return valid
}

type bufferDescriptorRing struct {
	bds   []*bufferDescriptor
	index int
	size  int
	stats *Stats
}

func (ring *bufferDescriptorRing) init(rx bool, n int, s *Stats) uint32 {
	ring.bds = make([]*bufferDescriptor, n)
	ring.size = n
	ring.stats = s

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
			Addr:  uint32(addr) + uint32(off),
			data:  data[off : off+dataSize],
			stats: ring.stats,
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
		bd.stats.EmptyFrames.Add(1)
		return
	}

	ring.next()

	if bd.Valid() {
		data = bd.Data()
	}

	// set empty
	bd.desc[3] |= (1 << BD_RX_ST_E) >> 8

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
