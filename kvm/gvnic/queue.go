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
)

const GVE_GQI_QPL_FORMAT = 0x02

const (
	ADMINQ_CREATE_TX_QUEUE = 0x2
	ADMINQ_CREATE_RX_QUEUE = 0x3
)

type queueResources struct {
	DBIndex      uint32
	CounterIndex uint32
	_            [56]byte
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
	// DMA buffers
	res  []byte
	desc []byte
	data []byte
}

type createRxQueueCommand struct {
	QueueID            uint32
	Index              uint32
	_                  uint32
	NtfyID             uint32
	QueueResourcesAddr uint64
	RxDescRingAddr     uint64
	RxDataRingAddr     uint64
	QueuePageListID    uint32
	RxRingSize         uint16
	PacketBufferSize   uint16
	RxBuffRingSize     uint16
	EnableRsc          uint8
	_                  [5]byte
}

func (hw *GVE) initRxQueue(index int) (err error) {
	var addr uint

	size := hw.Info.RxQueueEntries
	hw.rx = &rxQueue{}

	cmd := &createRxQueueCommand{
		QueueID:          uint32(index),
		Index:            uint32(index),
		NtfyID:           uint32(index + 1),
		QueuePageListID:  uint32(index + 1),
		RxRingSize:       size,
		PacketBufferSize: pageSize / 2,
	}

	// allocate queue resources
	n := binary.Size(queueResources{})
	addr, hw.rx.res = hw.Region.Reserve(n, 0)
	cmd.QueueResourcesAddr = uint64(addr)

	// allocate descriptor ring
	n = binary.Size(rxDesc{}) * int(size)
	addr, hw.rx.desc = hw.Region.Reserve(n, 0)
	cmd.RxDescRingAddr = uint64(addr)

	// allocate data ring
	n = 8 * int(size)
	addr, hw.rx.data = hw.Region.Reserve(n, 0)
	cmd.RxDataRingAddr = uint64(addr)

	// TODO: create Queue-page-list
	// RxPagesPerQpl (1024) * buffer
	//reg.Write(hw.rx.Doorbell, size)

	return hw.aq.Push(ADMINQ_CREATE_RX_QUEUE, cmd)
}
