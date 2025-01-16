// VirtIO Virtual Queue support
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
	"fmt"

	"github.com/usbarmory/tamago/dma"
)

// Descriptor Flags
const (
	Next     = 1
	Write    = 2
	Indirect = 3
)

// Ring represents a VirtIO virtual queue descriptor
type Descriptor struct {
	Address uint64
	Length  uint32
	Flags   uint16
	Next    uint16

	// DMA buffer
	buf []byte
}

// Bytes converts the descriptor structure to byte array format.
func (d *Descriptor) Bytes() []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, d.Address)
	binary.Write(buf, binary.LittleEndian, d.Length)
	binary.Write(buf, binary.LittleEndian, d.Flags)
	binary.Write(buf, binary.LittleEndian, d.Next)

	return buf.Bytes()
}

// Init initializes a virtual queue descriptor the given buffer length.
func (d *Descriptor) Init(length int, flags uint16) {
	addr, buf := dma.Reserve(length, 0)

	d.Address = uint64(addr)
	d.Length = uint32(length)
	d.Flags = flags

	d.buf = buf
}

// Destroy removes a virtual queue descriptor from physical memory.
func (d *Descriptor) Destroy() {
	dma.Release(uint(d.Address))
}

// Available represents a VirtIO virtual queue Available ring buffer
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

// Ring represents a VirtIO virtual queue buffer index
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

// Used represents a VirtIO virtual queue Used ring buffer
type Used struct {
	Flags      uint16
	Index      uint16
	Pad        [2]byte
	Ring       []*Ring
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

// VirtualQueue represents a VirtIO split virtual queue Descriptor
type VirtualQueue struct {
	Descriptors []*Descriptor
	Available   Available
	Used        Used

	// DMA buffer
	addr uint
	buf  []byte
}

// Bytes converts the descriptor structure to byte array format.
func (d *VirtualQueue) Bytes() []byte {
	buf := new(bytes.Buffer)

	for _, desc := range d.Descriptors {
		buf.Write(desc.Bytes())
	}

	buf.Write(d.Available.Bytes())
	buf.Write(d.Used.Bytes())

	return buf.Bytes()
}

// Init initializes a split virtual queue for the given size.
func (d *VirtualQueue) Init(size int, length int, flags uint16) {
	for i := 0; i < size; i++ {
		desc := &Descriptor{}
		desc.Init(length, flags)

		ring := &Ring{}

		d.Descriptors = append(d.Descriptors, desc)
		d.Available.Ring = append(d.Available.Ring, uint16(i))
		d.Used.Ring = append(d.Used.Ring, ring)
	}

	d.Available.Index = uint16(size)

	buf := d.Bytes()
	d.addr, d.buf = dma.Reserve(len(buf), 16)

	copy(d.buf, buf)
}

// Destroy removes a split virtual queue from physical memory.
func (d *VirtualQueue) Destroy() {
	for _, d := range d.Descriptors {
		d.Destroy()
	}

	dma.Release(d.addr)
}

// Address returns the virtual queue physical address.
func (d *VirtualQueue) Address() (desc uint, driver uint, device uint) {
	descSize := len((&Descriptor{}).Bytes()) * len(d.Descriptors)
	availSize := len(d.Available.Bytes())

	desc = d.addr
	driver = desc + uint(descSize)
	device = driver + uint(availSize)

	return
}

// Next makes the first available descriptor available for processing.
func (d *VirtualQueue) Next() {
	index := d.Available.Index

	d.Available.Ring[index] = index
	d.Available.Index += 1
}

func (d *VirtualQueue) Debug() {
	fmt.Printf("\n%+v\n", d)

	for _, desc := range d.Descriptors {
		fmt.Printf("%x\n", desc)
	}

	for _, ring := range d.Used.Ring {
		fmt.Printf("%x\n", ring)
	}
}
