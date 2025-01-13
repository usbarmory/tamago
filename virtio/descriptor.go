// VirtIO descriptor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package virtio

import (
	"bytes"
	"encoding/binary"
)

// VirtIO device types
const (
	NetworkCard = 0x01
)

// VirtIO MMIO Device Registers
const (
	Magic             = 0x000
	Version           = 0x004
	DeviceID          = 0x008
	VendorID          = 0x00c
	DeviceFeatures    = 0x010
	DeviceFeaturesSel = 0x014
	DriverFeatures    = 0x020
	DriverFeaturesSel = 0x024
	QueueSel          = 0x030
	QueueNumMax       = 0x034
	QueueNum          = 0x038
	QueueReady        = 0x044
	QueueNotify       = 0x050
	InterruptStatus   = 0x060
	InterruptACK      = 0x064
	Status            = 0x070
	QueueDesc         = 0x080
	QueueDriver       = 0x090
	QueueDevice       = 0x0a0
)

// VirtIO Device Status
const (
	DeviceAcknowledged = 0x01
	DriverLoaded       = 0x02
	DriverReady        = 0x03
	DeviceError        = 0x40
	DriverFailed       = 0x80
)

// Ring represents a VirtIO Virtual Queue buffer
type Buffer struct {
	Address uint64
	Length  uint32
	Flags   uint16
	Next    uint16
}

// Bytes converts the descriptor structure to byte array format.
func (d *Buffer) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, d)
	return buf.Bytes()
}

// Available represents a VirtIO Virtual Queue Available ring buffer
type Available struct {
	Flags      uint16
	Index      uint16
	Ring       []uint16
	EventIndex uint16
}

// Bytes converts the descriptor structure to byte array format.
func (d *Available) Bytes() []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, d.Flags)
	binary.Write(buf, binary.LittleEndian, d.Index)

	for _, ring := range d.Ring {
		binary.Write(buf, binary.LittleEndian, ring)
	}

	binary.Write(buf, binary.LittleEndian, d.EventIndex)

	return buf.Bytes()
}

// Ring represents a VirtIO Virtual Queue buffer index
type Ring struct {
	Index  uint32
	Length uint32
}

// Bytes converts the descriptor structure to byte array format.
func (d *Ring) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, d)
	return buf.Bytes()
}

// Used represents a VirtIO Virtual Queue Used ring buffer
type Used struct {
	Flags      uint16
	Index      uint16
	Pad        [2]byte
	Ring       []Ring
	AvailEvent uint16
}

// Bytes converts the descriptor structure to byte array format.
func (d *Used) Bytes() []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, d.Flags)
	binary.Write(buf, binary.LittleEndian, d.Index)
	binary.Write(buf, binary.LittleEndian, d.Pad)

	for _, ring := range d.Ring {
		buf.Write(ring.Bytes())
	}

	binary.Write(buf, binary.LittleEndian, d.AvailEvent)

	return buf.Bytes()
}

// VirtualQueue represents a VirtIO Virtual Queue Descriptor
type VirtualQueue struct {
	Buffers   []Buffer
	Available Available
	Used      Used
}

// Bytes converts the descriptor structure to byte array format.
func (d *VirtualQueue) Bytes() []byte {
	buf := new(bytes.Buffer)

	for _, buffer := range d.Buffers {
		buf.Write(buffer.Bytes())
	}

	buf.Write(d.Available.Bytes())
	buf.Write(make([]byte, 4096-buf.Len()))
	buf.Write(d.Used.Bytes())

	return buf.Bytes()
}
