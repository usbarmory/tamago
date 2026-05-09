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
	PageSize            uint64
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
	// control registers
	Doorbell uint32

	// Resources cache
	Resources *queueResources
	res       []byte // DMA buffer

	// DMA buffers
	desc        []byte
	data        []byte
	qpl         []byte
	qplPageList []byte
	qplListAddr uint
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
	// Resources cache
	Resources *queueResources

	// DMA buffers
	res         []byte
	desc        []byte
	qpl         []byte
	qplPageList []byte
	qplListAddr uint
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

func (hw *GVE) registerPageList(id int, size int) (addr uint, buf []byte, plAddr uint, pageList []byte, err error) {
	cmd := &registerPageListCommand{
		PageListID: uint32(id),
		NumPages:   uint32(size),
		PageSize:   pageSize,
	}

	// allocate queue pair list
	n := pageSize * int(cmd.NumPages)
	addr, buf = hw.Region.Reserve(n, pageSize)
	clear(buf)

	// allocate page list
	n = 8 * int(cmd.NumPages)
	plAddr, pageList = hw.Region.Reserve(n, pageSize)
	clear(pageList)
	cmd.PageAddressListAddr = uint64(plAddr)

	for i := uint64(0); i < uint64(cmd.NumPages); i++ {
		pga := uint64(addr) + i*pageSize
		binary.BigEndian.PutUint64(pageList[i*8:], pga)
	}

	if err = hw.aq.Push(ADMINQ_REGISTER_PAGE_LIST, cmd); err != nil {
		hw.Region.Release(plAddr)
		hw.Region.Release(addr)
		return 0, nil, 0, nil, err
	}

	// Keep the page-address list reserved until UNREGISTER_PAGE_LIST —
	// some firmware paths consult it while the QPL is active.
	return
}

func (hw *GVE) initRxQueue(id int) (err error) {
	var addr uint

	hw.rx = &rxQueue{}

	cmd := &createRxQueueCommand{
		QueueID: uint32(hw.Index),
		Index:   uint32(hw.Index),
		// NtfyID for RX queue 0 must equal num_tx_queues + 0 = 1 per
		// Linux gve_rx_idx_to_ntfy; ntfy_id 0 collides with TX queue 0
		// and the firmware silently drops inbound traffic on GCE N2D.
		NtfyID:           1,
		QueuePageListID:  uint32(id),
		RingSize:         hw.Info.RxQueueEntries,
		PacketBufferSize: pageSize / 2,
	}

	// Allocate descriptor/data rings before QPL registration, matching
	// Linux gve_alloc_queue_page_list ordering.
	hw.rx.Resources = &queueResources{}
	n := binary.Size(hw.rx.Resources)
	addr, hw.rx.res = hw.Region.Reserve(n, pageSize)
	clear(hw.rx.res)
	cmd.QueueResourcesAddr = uint64(addr)

	// allocate descriptor ring
	n = binary.Size(rxDesc{}) * int(cmd.RingSize)
	addr, hw.rx.desc = hw.Region.Reserve(n, pageSize)
	clear(hw.rx.desc)
	cmd.DescRingAddr = uint64(addr)

	// allocate data ring
	n = 8 * int(cmd.RingSize)
	addr, hw.rx.data = hw.Region.Reserve(n, pageSize)
	clear(hw.rx.data)
	cmd.DataRingAddr = uint64(addr)

	// fill data ring slots
	for i := uint64(0); i < uint64(cmd.RingSize); i++ {
		binary.BigEndian.PutUint64(hw.rx.data[i*8:], i*pageSize)
	}

	// register page list LAST (allocates the QPL pool + page-address list
	// internally and pushes the ADMINQ_REGISTER_PAGE_LIST command)
	size := int(hw.Info.RxPagesPerQpl)
	if _, hw.rx.qpl, hw.rx.qplListAddr, hw.rx.qplPageList, err = hw.registerPageList(id, size); err != nil {
		return
	}

	if err = hw.aq.Push(ADMINQ_CREATE_RX_QUEUE, cmd); err != nil {
		return
	}

	// get queue resources
	binary.Decode(hw.rx.res, binary.BigEndian, hw.rx.Resources)

	// get doorbell
	hw.rx.Doorbell = hw.doorbells + hw.rx.Resources.DBIndex*4

	// notify ring size
	cnt := bits.ReverseBytes32(uint32(hw.Info.RxQueueEntries))
	reg.Write(hw.rx.Doorbell, cnt)

	return
}

func (hw *GVE) initTxQueue(id int) (err error) {
	var addr uint

	hw.tx = &txQueue{
		Resources: &queueResources{},
	}

	cmd := &createTxQueueCommand{
		QueueID: uint32(hw.Index),
		// NtfyID for TX queue 0 is 0 per Linux gve_tx_idx_to_ntfy.
		NtfyID:          0,
		QueuePageListID: uint32(id),
		RingSize:        hw.Info.TxQueueEntries,
	}

	// Allocate queue resources and descriptor ring before QPL registration.
	n := binary.Size(queueResources{})
	addr, hw.tx.res = hw.Region.Reserve(n, pageSize)
	clear(hw.tx.res)
	cmd.QueueResourcesAddr = uint64(addr)

	n = binary.Size(txDesc{}) * int(cmd.RingSize)
	addr, hw.tx.desc = hw.Region.Reserve(n, pageSize)
	clear(hw.tx.desc)
	cmd.DescRingAddr = uint64(addr)

	size := int(hw.Info.TxPagesPerQpl)
	if _, hw.tx.qpl, hw.tx.qplListAddr, hw.tx.qplPageList, err = hw.registerPageList(id, size); err != nil {
		return
	}

	if err = hw.aq.Push(ADMINQ_CREATE_TX_QUEUE, cmd); err != nil {
		return
	}

	_, err = binary.Decode(hw.tx.res, binary.BigEndian, hw.tx.Resources)
	return
}
