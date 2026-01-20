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

const ADMINQ_CONFIGURE_DEVICE_RESOURCES = 0x2

type deviceResourcesCommand struct {
	CounterArray         uint64
	IRQDBAddr            uint64
	NumCounters          uint32
	NumIRQDBs            uint32
	IRQDBStride          uint32
	NtfyBlockMSIXBaseIdx uint32
	QueueFormat          uint8
	_                    [7]byte
}

type irqDoorbell struct {
	Index uint32
	_     [60]byte
}

func (hw *GVE) configureDeviceResources() (err error) {
	counterSize := int(hw.Info.Counters) * 4

	// allocate counter array
	counterArrayAddr, _ := hw.Region.Reserve(counterSize, pageSize)

	// allocate IRQ doorbell array
	doorbells := 2 // rx+tx
	irqDoorbellSize := binary.Size(irqDoorbell{})
	irqDBAddr, _ := hw.Region.Reserve(irqDoorbellSize*doorbells, 64)

	cmd := &deviceResourcesCommand{
		CounterArray:         uint64(counterArrayAddr),
		IRQDBAddr:            uint64(irqDBAddr),
		NumCounters:          uint32(hw.Info.Counters),
		NumIRQDBs:            uint32(doorbells),
		IRQDBStride:          uint32(irqDoorbellSize),
		NtfyBlockMSIXBaseIdx: 0,
		QueueFormat:          GVE_GQI_QPL_FORMAT,
	}

	return hw.aq.Push(ADMINQ_CONFIGURE_DEVICE_RESOURCES, cmd)
}
