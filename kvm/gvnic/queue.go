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
	"errors"
	"math/bits"

	"github.com/usbarmory/tamago/internal/reg"
)

const (
	ADMINQ_REGISTER_PAGE_LIST = 0x3
	ADMINQ_CREATE_TX_QUEUE    = 0x5
	ADMINQ_CREATE_RX_QUEUE    = 0x6

	GVE_GQI_QPL_FORMAT = 0x02
	GVE_TXD_STD        = 0x00
)

const (
	flagsMask = 0b111
	rxPadLen  = 2
)

type queueResources struct {
	DBIndex      uint32
	CounterIndex uint32
	_            [56]byte
}

type registerPageListCommand struct {
	PageListID          uint32
	NumPages            uint32
	PageAddressListAddr uint64
	_                   [16]byte
}

// TODO: convert to offsets
type rxDesc struct {
	_        [48]byte
	RSSHash  uint32
	MSS      uint16
	_        uint16
	HdrLen   uint8
	HdrOff   uint8
	Csum     uint16
	Len      uint16
	FlagsSeq uint16
}

type queue struct {
	// control registers
	Doorbell uint32

	// Resources cache
	Resources *queueResources
	res       []byte // DMA buffer

	// DMA buffers
	desc []byte
	data []byte
	qpl  []byte

	// ring state
	size uint32
}

func (q *queue) setDoorbell(base uint32) {
	// get queue resources
	binary.Decode(q.res, binary.BigEndian, q.Resources)

	// get doorbell
	q.Doorbell = base + q.Resources.DBIndex*4
}

type rxQueue struct {
	queue

	// ring state
	cnt   uint32
	fill  uint32
	seqno uint16
}

func (rx *rxQueue) next() {
	// advance the read cursor
	rx.cnt++

	if rx.seqno == flagsMask {
		rx.seqno = 1
	} else {
		rx.seqno += 1
	}

	// recycle consumed buffers when half is drained
	if rx.fill-rx.cnt > rx.size/2 {
		return
	}

	rx.fill = rx.cnt + rx.size

	// notify ring size
	n := bits.ReverseBytes32(rx.fill)
	reg.Write(rx.Doorbell, n)
}

type createRxQueueCommand struct {
	QueueID            uint32
	Index              uint32
	_                  uint32
	NtfyID             uint32
	QueueResourcesAddr uint64
	DescRingAddr       uint64
	DataRingAddr       uint64
	QueuePageListID    uint32
	RingSize           uint16
	PacketBufferSize   uint16
	BuffRingSize       uint16
	EnableRsc          uint8
	_                  [5]byte
}

// TODO: convert to offsets
type txDesc struct {
	TypeFlags uint8
	CsumOff   uint8
	HdrOff    uint8
	DescCnt   uint8
	Len       uint16
	SegLen    uint16
	SegAddr   uint64
}

type txQueue struct {
	queue

	// ring state
	head  uint32
	tail  uint32
}

func (tx *txQueue) next() {
	tx.head++

	// notify ring size
	n := bits.ReverseBytes32(tx.head)
	reg.Write(tx.Doorbell, n)
}

type createTxQueueCommand struct {
	QueueID            uint32
	_                  uint32
	QueueResourcesAddr uint64
	DescRingAddr       uint64
	QueuePageListID    uint32
	NtfyID             uint32
	CompRingAddr       uint64
	RingSize           uint16
	CompRingSize       uint16
	_                  [4]byte
}

func (hw *GVE) registerPageList(id int, size int) (addr uint, buf []byte, err error) {
	cmd := &registerPageListCommand{
		PageListID: uint32(id),
		NumPages:   uint32(size),
	}

	// allocate queue pair list
	n := pageSize * int(cmd.NumPages)
	addr, buf = hw.Region.Reserve(n, pageSize)

	// allocate page list
	n = 8 * int(cmd.NumPages)
	plAddr, pageList := hw.Region.Reserve(n, pageSize)
	cmd.PageAddressListAddr = uint64(plAddr)
	defer hw.Region.Release(plAddr)

	for i := uint64(0); i < uint64(cmd.NumPages); i++ {
		pga := uint64(addr) + i*pageSize
		binary.BigEndian.PutUint64(pageList[i*8:], pga)
	}

	if err = hw.aq.Push(ADMINQ_REGISTER_PAGE_LIST, cmd); err != nil {
		return 0, nil, err
	}

	return
}

func (hw *GVE) initRxQueue(id int) (err error) {
	var addr uint

	queueSize := uint32(hw.Info.RxQueueEntries)
	qplSize := int(hw.Info.RxPagesPerQpl)

	hw.rx = &rxQueue{
		cnt:   0,
		fill:  queueSize,
		seqno: 1,
	}
	hw.rx.size = hw.rx.fill

	if _, hw.rx.qpl, err = hw.registerPageList(id, qplSize); err != nil {
		return
	}

	cmd := &createRxQueueCommand{
		QueueID:          uint32(hw.Index),
		Index:            uint32(hw.Index),
		NtfyID:           uint32(id),
		QueuePageListID:  uint32(id),
		RingSize:         uint16(hw.rx.size),
		PacketBufferSize: pageSize / 2,
	}

	// allocate queue resources
	hw.rx.Resources = &queueResources{}
	n := binary.Size(hw.rx.Resources)
	addr, hw.rx.res = hw.Region.Reserve(n, 64)
	cmd.QueueResourcesAddr = uint64(addr)

	// allocate descriptor ring
	n = binary.Size(rxDesc{}) * int(cmd.RingSize)
	addr, hw.rx.desc = hw.Region.Reserve(n, 0)
	cmd.DescRingAddr = uint64(addr)

	// zero out descriptor ring contents
	for i := range hw.rx.desc {
		hw.rx.desc[i] = 0x00
	}

	// allocate data ring
	n = 8 * int(cmd.RingSize)
	addr, hw.rx.data = hw.Region.Reserve(n, 0)
	cmd.DataRingAddr = uint64(addr)

	// fill data ring slots
	for i := uint64(0); i < uint64(cmd.RingSize); i++ {
		binary.BigEndian.PutUint64(hw.rx.data[i*8:], i*pageSize)
	}

	if err = hw.aq.Push(ADMINQ_CREATE_RX_QUEUE, cmd); err != nil {
		return
	}

	hw.rx.setDoorbell(hw.doorbells)

	// notify ring size
	cnt := bits.ReverseBytes32(queueSize)
	reg.Write(hw.rx.Doorbell, cnt)

	return
}

func (hw *GVE) initTxQueue(id int) (err error) {
	var addr uint

	queueSize := uint32(hw.Info.TxQueueEntries)
	qplSize := int(hw.Info.TxPagesPerQpl)

	hw.tx = &txQueue{}
	hw.tx.size = queueSize

	if _, hw.tx.qpl, err = hw.registerPageList(id, qplSize); err != nil {
		return
	}

	cmd := &createTxQueueCommand{
		QueueID:         uint32(hw.Index),
		NtfyID:          uint32(id),
		QueuePageListID: uint32(id),
		RingSize:        hw.Info.TxQueueEntries,
	}

	// allocate queue resources
	hw.tx.Resources = &queueResources{}
	n := binary.Size(hw.tx.Resources)
	addr, hw.tx.res = hw.Region.Reserve(n, 64)
	cmd.QueueResourcesAddr = uint64(addr)

	// allocate descriptor ring
	n = binary.Size(txDesc{}) * int(cmd.RingSize)
	addr, hw.tx.desc = hw.Region.Reserve(n, 0)
	cmd.DescRingAddr = uint64(addr)

	if err = hw.aq.Push(ADMINQ_CREATE_TX_QUEUE, cmd); err != nil {
		return
	}

	hw.tx.setDoorbell(hw.doorbells)

	return
}

func (hw *GVE) Receive(buf []byte) (n int, err error) {
	if len(buf) == 0 {
		return
	}

	idx := hw.rx.cnt % hw.rx.size
	off := uint(idx) * 64

	flagsSeq := binary.BigEndian.Uint16(hw.rx.desc[off+62:]) // rxDesc.FlagsSeq
	length := binary.BigEndian.Uint16(hw.rx.desc[off+60:])   // rxDesc.Len

	if flagsSeq&flagsMask != hw.rx.seqno {
		return 0, nil
	}

	defer hw.rx.next()

	if length <= rxPadLen {
		return 0, nil
	}

	// the data ring holds 64-bit QPL offsets pointing to the actual data
	qplOff := uint(binary.BigEndian.Uint64(hw.rx.data[idx*8:]))
	data := hw.rx.qpl[qplOff+rxPadLen : qplOff+uint(length)]

	n = copy(buf, data)

	return n, nil
}

func (hw *GVE) Transmit(buf []byte) (err error) {
	if len(buf) > pageSize {
		return errors.New("frame too large")
	}

	txPages := uint32(hw.Info.TxPagesPerQpl)
	idx := hw.tx.head % hw.tx.size
	qplOff := (hw.tx.head % txPages) * pageSize

	cntIndex := hw.tx.Resources.CounterIndex * 4
	hw.tx.tail = binary.BigEndian.Uint32(hw.counters[cntIndex:])

	inflight := hw.tx.head - hw.tx.tail

	if inflight >= hw.tx.size || inflight >= txPages {
		return errors.New("tx queue full")
	}

	// copy the frame into the TX QPL
	copy(hw.tx.qpl[qplOff:qplOff+uint32(len(buf))], buf)

	// write one standard descriptor
	off := uint(idx) * uint(binary.Size(txDesc{}))
	d := hw.tx.desc
	d[off+0] = GVE_TXD_STD                                  // type_flags
	d[off+1] = 0                                            // l4_csum_offset
	d[off+2] = 0                                            // l4_hdr_offset
	d[off+3] = 1                                            // desc_cnt
	binary.BigEndian.PutUint16(d[off+4:], uint16(len(buf))) // len
	binary.BigEndian.PutUint16(d[off+6:], uint16(len(buf))) // seg_len
	binary.BigEndian.PutUint64(d[off+8:], uint64(qplOff))   // seg_addr

	hw.tx.next()

	return nil
}
