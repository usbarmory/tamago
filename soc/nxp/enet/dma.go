// NXP 10/100-Mbps Ethernet MAC (ENET)
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package enet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/internal/reg"
)

const (
	MTU             = 1500
	maxFrameSize    = MTU + 18
	minFrameSize    = 42
	defaultRingSize = 16

	// On arm64 Go `copy` built-in cannot be safely used on device memory
	// as LDP/STP instructions require 8-byte alignment, for this reason
	// all `copy` against DMA buffers are forced to 8-byte aligned slices.
	copyAlign   = 8
	bufferAlign = 64
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

func (bd *bufferDescriptor) Read(buf []byte) (n int, err error) {
	n = int(bd.Length) - 4

	if len(buf) < n {
		return 0, fmt.Errorf("buffer too small (%d < %d)", len(buf), n)
	}

	r := n % copyAlign

	if r != 0 {
		n -= r
	}

	copy(buf[:n], bd.data[:n])

	if r != 0 {
		for i := range r {
			buf[n+i] = bd.data[n+i]
		}
	}

	return
}

func (bd *bufferDescriptor) Valid() bool {
	s := uint32(bd.Status)

	switch {
	case s&(1<<BD_ST_L) == 0:
		return false
	case s&frameErrorMask != 0:
		if (s>>BD_RX_ST_OV)&1 == 1 {
			bd.stats.Overrun += 1
		} else {
			bd.stats.FrameLengthViolation += (s >> BD_RX_ST_LG) & 1
			bd.stats.NonOctetAlignedFrame += (s >> BD_RX_ST_NO) & 1
			bd.stats.CRCOrFrameError += (s >> BD_RX_ST_CR) & 1
		}

		return false
	case bd.Length < minFrameSize:
		bd.stats.FrameTooSmall += 1
		return false
	case bd.Length > maxFrameSize:
		bd.stats.FrameTooLarge += 1
		return false
	}

	return true
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

	// To avoid excessive DMA region fragmentation, a single allocation
	// reserves all descriptors and data pointers.

	descSize := len((&bufferDescriptor{}).Bytes())
	ptr, desc := dma.Reserve(n*descSize, bufferAlign)

	dataSize := maxFrameSize + (bufferAlign - (maxFrameSize % bufferAlign))
	addr, data := dma.Reserve(n*dataSize, bufferAlign)

	for i := range n {
		off := dataSize * i

		bd := &bufferDescriptor{
			Addr:  uint32(addr) + uint32(off),
			data:  data[off : off+dataSize],
			stats: s,
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

func (ring *bufferDescriptorRing) pop(buf []byte) (n int, err error) {
	bd := ring.bds[ring.index]

	bd.Length = uint16(bd.desc[0])
	bd.Length |= uint16(bd.desc[1]) << 8

	bd.Status = uint16(bd.desc[2])
	bd.Status |= uint16(bd.desc[3]) << 8

	if bd.Status&(1<<BD_RX_ST_E) != 0 {
		return
	}

	ring.next()

	if bd.Valid() {
		n, err = bd.Read(buf)
	}

	// set empty
	bd.desc[3] |= (1 << BD_RX_ST_E) >> 8

	return
}

func (ring *bufferDescriptorRing) push(data []byte) (err error) {
	bd := ring.bds[ring.index]

	if uint16(bd.desc[3]<<8)&(1<<BD_TX_ST_R) != 0 {
		return errors.New("frame not sent")
	}

	bd.Length = uint16(len(data))
	bd.Status = (1 << BD_ST_L) | (1 << BD_TX_ST_TC)

	bd.desc[0] = byte(bd.Length & 0xff)
	bd.desc[1] = byte((bd.Length & 0xff00) >> 8)

	bd.desc[2] = byte((bd.Status & 0xff))
	bd.desc[3] = byte((bd.Status & 0xff00) >> 8)

	if r := len(data) % copyAlign; r != 0 {
		data = append(data, make([]byte, copyAlign-r)...)
	}

	copy(bd.data, data)

	if ring.next() {
		bd.desc[3] |= (1 << BD_ST_W) >> 8
	}

	// set ready
	bd.desc[3] |= (1 << BD_TX_ST_R) >> 8

	return
}

// Receive receives a single Ethernet frame, excluding the checksum, from the
// MAC controller ring buffer.
func (hw *ENET) Receive(buf []byte) (n int, err error) {
	n, err = hw.rx.pop(buf)
	reg.Set(hw.rdar, RDAR_ACTIVE)
	return
}

// Transmit transmits a single Ethernet frame, the checksum is appended
// automatically and must not be included.
func (hw *ENET) Transmit(buf []byte) (err error) {
	if len(buf) > maxFrameSize {
		return errors.New("frame too large")
	}

	err = hw.tx.push(buf)
	reg.Set(hw.tdar, TDAR_ACTIVE)
	return
}
