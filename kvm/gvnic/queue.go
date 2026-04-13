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
	ADMINQ_REGISTER_PAGE_LIST = 0x3
	ADMINQ_CREATE_TX_QUEUE    = 0x5
	ADMINQ_CREATE_RX_QUEUE    = 0x6

	GVE_GQI_QPL_FORMAT = 0x02

	rxID = 0
	txID = 1
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

type rxQueue struct {
	Resources *queueResources

	// DMA buffers
	res  []byte
	desc []byte
	data []byte

	qplAddr uint
	qpl     []byte
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

type txDesc struct {
	TypeFlags uint8
	CsumOff   uint8
	HdrOff    uint8
	DescCnt   uint8
	Len       uint16
	SegLen    uint16
	SeqAddr   uint64 
}

type txQueue struct {
	// DMA buffers
	res  []byte
	desc []byte

	qplAddr uint
	qpl     []byte
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

	hw.rx = &rxQueue{}
	size := int(hw.Info.RxPagesPerQpl)

	if hw.rx.qplAddr, hw.rx.qpl, err = hw.registerPageList(id, size); err != nil {
		return
	}

	cmd := &createRxQueueCommand{
		QueueID:          uint32(hw.Index),
		Index:            uint32(hw.Index),
		NtfyID:           uint32(id),
		QueuePageListID:  uint32(id),
		RingSize:         hw.Info.RxQueueEntries,
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

	// allocate data ring
	n = 8 * int(cmd.RingSize)
	addr, hw.rx.data = hw.Region.Reserve(n, 0)
	cmd.DataRingAddr = uint64(addr)

	for i := uint64(0); i < uint64(cmd.RingSize); i++ {
		binary.BigEndian.PutUint64(hw.rx.data[i*8:], i*pageSize)
	}

	if err = hw.aq.Push(ADMINQ_CREATE_RX_QUEUE, cmd); err != nil {
		return
	}

	binary.Decode(hw.rx.res, binary.BigEndian, hw.rx.Resources)

	off := hw.rx.Resources.DBIndex * 4
	cnt := bits.ReverseBytes32(uint32(hw.Info.RxQueueEntries))

	reg.Write(hw.doorbells + off, cnt)

	return
}

func (hw *GVE) initTxQueue(id int) (err error) {
	var addr uint

	hw.tx = &txQueue{}
	size := int(hw.Info.TxPagesPerQpl)

	if hw.tx.qplAddr, hw.tx.qpl, err = hw.registerPageList(id, size); err != nil {
		return
	}

	cmd := &createTxQueueCommand{
		QueueID:          uint32(hw.Index),
		NtfyID:           uint32(id),
		QueuePageListID:  uint32(id),
		RingSize:         hw.Info.TxQueueEntries,
	}

	// allocate queue resources
	n := binary.Size(queueResources{})
	addr, hw.tx.res = hw.Region.Reserve(n, 64)
	cmd.QueueResourcesAddr = uint64(addr)

	// allocate descriptor ring
	n = binary.Size(txDesc{}) * int(cmd.RingSize)
	addr, hw.tx.desc = hw.Region.Reserve(n, 0)
	cmd.DescRingAddr = uint64(addr)

	return hw.aq.Push(ADMINQ_CREATE_TX_QUEUE, cmd)
}
