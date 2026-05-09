// Google Compute Engine Virtual Ethernet (gVNIC) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package gvnic

import (
	"encoding/binary"
	"math/bits"

	"github.com/usbarmory/tamago/internal/reg"
)

const (
	// GQI-QPL uses the registered page list as a byte-addressed FIFO.
	txFIFOAlign       = 64
	txMinPktDescBytes = 182
	txMaxAllocFrags   = 2
)

// txFIFOState tracks the FIFO head pointer and available space in the TX QPL.
type txFIFOState struct {
	head  uint32
	avail uint32
	size  uint32
}

// txBufferState records per-descriptor FIFO consumption for TX completion.
type txBufferState struct {
	fifoBytes uint32
	frameLen  uint16
}

type txIovec struct {
	off     uint32
	length  uint32
	padding uint32
}

// Transmit sends a single Ethernet frame. Satisfies go-net NetworkDevice.
func (hw *GVE) Transmit(buf []byte) error {
	if hw.tx == nil || len(buf) == 0 {
		return nil
	}

	hw.Lock()
	defer hw.Unlock()

	hw.cleanTxDone()

	n := len(buf)
	fifo := &hw.state.txFIFO
	if n > int(fifo.size) {
		return nil
	}

	pad := fifo.txFIFOPadAllocOneFrag(n)
	if pad >= txMinPktDescBytes {
		pad = 0
	}
	iovi := 0
	if pad > 0 {
		iovi = 1
	}
	nfrags := fifo.previewAllocFrags(pad + n)
	ndescs := nfrags - iovi
	if ndescs <= 0 {
		return nil
	}

	mask := hw.state.txMask
	if !hw.canPostTx(ndescs, pad+n) {
		hw.cleanTxDone()
		if !hw.canPostTx(ndescs, pad+n) {
			return nil
		}
	}

	var iov [txMaxAllocFrags]txIovec
	nfrags = fifo.txAllocFIFO(pad+n, &iov)
	ndescs = nfrags - iovi

	start := hw.state.txReq
	reqi := hw.state.txReq
	copyOff := 0
	var spaceUsed uint32

	for i := iovi; i < nfrags; i++ {
		spaceUsed += iov[i].length + iov[i].padding
		descOff := int(reqi&mask) * txDescSize
		desc := hw.tx.desc[descOff : descOff+txDescSize]

		if copyOff == 0 {
			fillTxPktDesc(desc, uint8(ndescs), uint16(iov[i].length), iov[i].off, uint16(n))
		} else {
			fillTxSegDesc(desc, uint16(iov[i].length), iov[i].off)
		}

		copy(hw.tx.qpl[iov[i].off:iov[i].off+iov[i].length], buf[copyOff:copyOff+int(iov[i].length)])
		copyOff += int(iov[i].length)
		reqi++
	}

	hw.state.txPending[start&mask] = txBufferState{
		fifoBytes: spaceUsed,
		frameLen:  uint16(n),
	}
	for i := uint32(1); i < uint32(ndescs); i++ {
		hw.state.txPending[(start+i)&mask] = txBufferState{}
	}

	hw.state.txReq += uint32(ndescs)

	hw.ringTxDoorbell()
	return nil
}

func (hw *GVE) canPostTx(descs int, bytes int) bool {
	avail := hw.state.txMask + 1 - (hw.state.txReq - hw.state.txDone)
	return avail >= uint32(descs) && hw.state.txFIFO.avail > uint32(bytes)
}

func (f *txFIFOState) txFIFOPadAllocOneFrag(bytes int) int {
	if f.head+uint32(bytes) < f.size {
		return 0
	}
	return int(f.size - f.head)
}

func (f *txFIFOState) previewAllocFrags(bytes int) int {
	if bytes == 0 {
		return 0
	}
	if f.head+uint32(bytes) > f.size {
		return 2
	}
	return 1
}

func (f *txFIFOState) txAllocFIFO(bytes int, iov *[txMaxAllocFrags]txIovec) int {
	if bytes == 0 {
		return 0
	}

	nfrags := 1
	iov[0] = txIovec{off: f.head, length: uint32(bytes)}
	f.head += uint32(bytes)

	if f.head > f.size {
		nfrags = 2
		overflow := f.head - f.size
		iov[0].length -= overflow
		iov[1] = txIovec{off: 0, length: overflow}
		f.head = overflow
	}

	alignedHead := (f.head + (txFIFOAlign - 1)) &^ (txFIFOAlign - 1)
	padding := alignedHead - f.head
	iov[nfrags-1].padding = padding
	f.avail -= uint32(bytes) + padding
	f.head = alignedHead
	if f.head == f.size {
		f.head = 0
	}

	return nfrags
}

func fillTxPktDesc(desc []byte, descCnt uint8, segLen uint16, segAddr uint32, pktLen uint16) {
	desc[0] = GVE_TXD_STD
	desc[1] = 0
	desc[2] = 0
	desc[3] = descCnt
	binary.BigEndian.PutUint16(desc[4:6], pktLen)
	binary.BigEndian.PutUint16(desc[6:8], segLen)
	binary.BigEndian.PutUint64(desc[8:16], uint64(segAddr))
}

func fillTxSegDesc(desc []byte, segLen uint16, segAddr uint32) {
	desc[0] = GVE_TXD_SEG
	desc[1] = 0
	binary.BigEndian.PutUint16(desc[2:4], 0)
	binary.BigEndian.PutUint16(desc[4:6], 0)
	binary.BigEndian.PutUint16(desc[6:8], segLen)
	binary.BigEndian.PutUint64(desc[8:16], uint64(segAddr))
}

func (hw *GVE) cleanTxDone() {
	if hw.tx == nil || hw.tx.Resources == nil || hw.state.counterArray == nil {
		return
	}

	mask := hw.state.txMask

	off := int(hw.tx.Resources.CounterIndex) * 4
	if off+4 > len(hw.state.counterArray) {
		return
	}
	nicDone := binary.BigEndian.Uint32(hw.state.counterArray[off : off+4])
	if nicDone == hw.state.txDone || nicDone > hw.state.txReq {
		return
	}

	for hw.state.txDone != nicDone {
		idx := hw.state.txDone & mask
		info := &hw.state.txPending[idx]
		if info.fifoBytes != 0 {
			hw.state.txFIFO.avail += info.fifoBytes
			*info = txBufferState{}
		}
		hw.state.txDone++
	}
}

func (hw *GVE) ringTxDoorbell() {
	if hw.tx == nil || hw.tx.Resources == nil {
		return
	}
	reg.Write(hw.doorbells+hw.tx.Resources.DBIndex*4, bits.ReverseBytes32(hw.state.txReq))
}
