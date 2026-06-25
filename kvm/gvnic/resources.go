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
	ADMINQ_CONFIGURE_DEVICE_RESOURCES = 0x2

	irqDoorbells = 2 // rx+tx
)

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

func (hw *GVE) configureDeviceResources() (err error) {
	// allocate counter array
	counterSize := int(hw.Info.Counters) * 4
	counterArrayAddr, counters := hw.Region.Reserve(counterSize, pageSize)

	// allocate IRQ doorbells array
	irqDBAddr, irqs := hw.Region.Reserve(descSize*irqDoorbells, pageSize)

	// zero out DMA pages
	clear(counters)
	clear(irqs)

	hw.counters = counters
	hw.irqs = irqs

	cmd := &deviceResourcesCommand{
		CounterArray:         uint64(counterArrayAddr),
		IRQDBAddr:            uint64(irqDBAddr),
		NumCounters:          uint32(hw.Info.Counters),
		NumIRQDBs:            irqDoorbells,
		IRQDBStride:          uint32(descSize),
		NtfyBlockMSIXBaseIdx: 0,
		QueueFormat:          GVE_GQI_QPL_FORMAT,
	}

	if err = hw.aq.Push(ADMINQ_CONFIGURE_DEVICE_RESOURCES, cmd); err != nil {
		return
	}

	for i := range irqDoorbells {
		off := i * descSize
		dbIdx := binary.BigEndian.Uint32(hw.irqs[off : off+4])
		reg.Write(hw.doorbells+dbIdx*4, bits.ReverseBytes32(0))
	}

	return
}
