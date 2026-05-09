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
	if counterSize < pageSize {
		counterSize = pageSize
	}

	// allocate counter array
	counterArrayAddr, counterBuf := hw.Region.Reserve(counterSize, pageSize)
	clear(counterBuf)
	hw.state.counterArray = counterBuf

	// allocate IRQ doorbell array
	doorbells := 2 // rx+tx
	irqDoorbellSize := binary.Size(irqDoorbell{})
	irqDBBytes := irqDoorbellSize * doorbells
	if irqDBBytes < pageSize {
		irqDBBytes = pageSize
	}
	irqDBAddr, irqDBBuf := hw.Region.Reserve(irqDBBytes, pageSize)
	clear(irqDBBuf)

	// Stash the IRQ DB array + count so unmaskAllIRQs() (called after
	// queue creation) can read the device-written ntfy_block doorbell
	// indices and write BE32 0 to each, unmasking notifications. Without
	// the stash + the unmask the firmware holds inbound traffic.
	hw.state.irqDBArray = irqDBBuf
	hw.state.numIRQDBs = uint32(doorbells)

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

// unmaskAllIRQs writes BE32 0 to each notification block's doorbell slot
// in BAR2, unmasking notifications (Linux gve_turnup). Without this the
// firmware holds inbound traffic even when the driver is polling.
//
// The IRQ DB array layout: numIRQDBs entries, each irqDBStride bytes,
// where bytes [0:4] of each entry hold the device-written ntfy_block
// doorbell index (BE32). The doorbell write target is BAR2 + dbIdx*4.
func (hw *GVE) unmaskAllIRQs() {
	for i := 0; i < int(hw.state.numIRQDBs); i++ {
		off := i * irqDBStride
		if off+4 > len(hw.state.irqDBArray) {
			break
		}
		dbIdx := binary.BigEndian.Uint32(hw.state.irqDBArray[off : off+4])
		reg.Write(hw.doorbells+dbIdx*4, bits.ReverseBytes32(0))
	}
}
