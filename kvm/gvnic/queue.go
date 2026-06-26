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

	irqAck   = 1 << 31
	irqEvent = 1 << 29
)

// TX descriptor offsets
const (
	txDescSize  = 16
	txTypeFlags = 0
	txCsumOff   = 1
	txHdrOff    = 2
	txDescCnt   = 3
	txLen       = 4
	txSegLen    = 6
	txSegAddr   = 8
)

// RX descriptor offsets
const (
	rxDescSize = descSize
	rxLen      = 60
	rxFlagsSeq = 62
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

type queue struct {
	id int

	// control registers
	Doorbell    uint32
	DoorbellIRQ uint32

	// Resources cache
	Resources *queueResources

	// DMA buffers
	res  []byte
	desc []byte
	data []byte
	qpl  []byte

	// ring state
	size uint32
}

func (q *queue) init(hw *GVE, ringSize uint16) (resAddr, descAddr uint64) {
	var addr uint

	// allocate queue resources
	q.Resources = &queueResources{}
	n := binary.Size(q.Resources)
	addr, q.res = hw.Region.Reserve(n, pageSize)
	resAddr = uint64(addr)

	// allocate descriptor ring
	n = rxDescSize * int(ringSize)
	addr, q.desc = hw.Region.Reserve(n, pageSize)
	descAddr = uint64(addr)

	// zero out DMA pages
	clear(q.res)
	clear(q.desc)

	return
}

func (q *queue) setDoorbells(hw *GVE) {
	binary.Decode(q.res, binary.BigEndian, q.Resources)
	q.Doorbell = hw.doorbells + q.Resources.DBIndex*4

	irqIndex := binary.BigEndian.Uint32(hw.irqs[q.id*descSize:])
	q.DoorbellIRQ = hw.doorbells + irqIndex*4
}

func (q *queue) ack() {
	// ack IRQ, re-enable event delivery
	reg.Write(q.DoorbellIRQ, bits.ReverseBytes32(irqAck|irqEvent))
}

type txQueue struct {
	queue

	// ring state
	head uint32
	tail uint32
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

	// zero out DMA pages
	clear(buf)
	clear(pageList)

	for i := uint64(0); i < uint64(cmd.NumPages); i++ {
		pga := uint64(addr) + i*pageSize
		binary.BigEndian.PutUint64(pageList[i*8:], pga)
	}

	if err = hw.aq.Push(ADMINQ_REGISTER_PAGE_LIST, cmd); err != nil {
		return 0, nil, err
	}

	return
}

func (hw *GVE) initTxQueue(id int) (err error) {
	queueSize := uint32(hw.Info.TxQueueEntries)
	qplSize := int(hw.Info.TxPagesPerQpl)

	hw.tx = &txQueue{}
	hw.tx.id = id
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

	cmd.QueueResourcesAddr, cmd.DescRingAddr = hw.tx.init(hw, cmd.RingSize)

	if err = hw.aq.Push(ADMINQ_CREATE_TX_QUEUE, cmd); err != nil {
		return
	}

	hw.tx.setDoorbells(hw)

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

	hw.rx.id = id
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

	cmd.QueueResourcesAddr, cmd.DescRingAddr = hw.rx.init(hw, cmd.RingSize)

	// allocate data ring
	n := 8 * int(cmd.RingSize)
	addr, hw.rx.data = hw.Region.Reserve(n, pageSize)
	cmd.DataRingAddr = uint64(addr)

	// fill data ring slots
	for i := uint64(0); i < uint64(cmd.RingSize); i++ {
		binary.BigEndian.PutUint64(hw.rx.data[i*8:], i*pageSize)
	}

	if err = hw.aq.Push(ADMINQ_CREATE_RX_QUEUE, cmd); err != nil {
		return
	}

	hw.rx.setDoorbells(hw)

	// notify ring size
	cnt := bits.ReverseBytes32(queueSize)
	reg.Write(hw.rx.Doorbell, cnt)

	return
}

func (hw *GVE) Receive(buf []byte) (n int, err error) {
	if len(buf) == 0 {
		return
	}

	idx := hw.rx.cnt % hw.rx.size
	off := uint(idx) * rxDescSize

	length := binary.BigEndian.Uint16(hw.rx.desc[off+rxLen:])
	flagsSeq := binary.BigEndian.Uint16(hw.rx.desc[off+rxFlagsSeq:])

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

	off := uint(idx) * txDescSize
	tx := hw.tx.desc

	tx[off+txTypeFlags] = GVE_TXD_STD
	tx[off+txCsumOff] = 0
	tx[off+txHdrOff] = 0
	tx[off+txDescCnt] = 1

	binary.BigEndian.PutUint16(tx[off+txLen:], uint16(len(buf)))
	binary.BigEndian.PutUint16(tx[off+txSegLen:], uint16(len(buf)))
	binary.BigEndian.PutUint64(tx[off+txSegAddr:], uint64(qplOff))

	hw.tx.next()

	return nil
}
