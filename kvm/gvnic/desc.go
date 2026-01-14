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

	"github.com/usbarmory/tamago/dma"
)

const (
	ADMINQ_DESCRIBE_DEVICE           = 0x1
	ADMINQ_DEVICE_DESCRIPTOR_VERSION = 1
)

type deviceDescriptorCommand struct {
	Address uint64
	Version uint32
	Length  uint32
}

type DeviceDescriptor struct {
	MaxRegisteredPages uint64
	_                  uint16
	TxQueueEntries     uint16
	RxQueueEntries     uint16
	DefaultNumQueues   uint16
	MTU                uint16
	Counters           uint16
	TxPagesPerQpl      uint16
	RxPagesPerQpl      uint16
	MAC                [6]byte
	NumDeviceOptions   uint16
	TotalLength        uint16
	_                  [6]byte
}

func (hw *GVE) describeDevice() (err error) {
	if hw.Info == nil {
		hw.Info = &DeviceDescriptor{}
	}

	addr, buf := dma.Reserve(pageSize, pageSize)
	defer dma.Release(addr)

	cmd := &deviceDescriptorCommand{
		Address: uint64(addr),
		Version: ADMINQ_DEVICE_DESCRIPTOR_VERSION,
		Length:  pageSize,
	}

	if err = hw.aq.Push(ADMINQ_DESCRIBE_DEVICE, cmd); err != nil {
		return
	}

	_, err = binary.Decode(buf, binary.BigEndian, hw.Info)

	return
}
