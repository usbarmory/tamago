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
	"sync"

	"github.com/usbarmory/tamago/dma"
)

// Descriptor Flags
const (
	Next     = 1
	Write    = 2
	Indirect = 3
)

// Descriptor represents a VirtIO virtual queue descriptor.
//
// All exported fields are used one-time at initialization, fields requiring
// DMA are accessible through functions.
type Descriptor struct {
	Address uint64
	length  uint32
	Flags   uint16
	Next    uint16

	// DMA buffer
	buf []byte
}

// Bytes converts the descriptor structure to byte array format.
func (d *Descriptor) Bytes() []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, d.Address)
	binary.Write(buf, binary.LittleEndian, d.length)
	binary.Write(buf, binary.LittleEndian, d.Flags)
	binary.Write(buf, binary.LittleEndian, d.Next)

	return buf.Bytes()
}

// Length updates the descriptor length field.
func (d *Descriptor) Length(length uint32) {
	off := 8
	binary.LittleEndian.PutUint32(d.buf[off:], length)

	d.length = length
}

// Init initializes a virtual queue descriptor the given buffer length.
func (d *Descriptor) Init(length int, flags uint16) {
	addr, buf := dma.Reserve(length, 0)

	d.Address = uint64(addr)
	d.length = uint32(length)
	d.Flags = flags

	d.buf = buf
}

// Destroy removes a virtual queue descriptor from physical memory.
func (d *Descriptor) Destroy() {
	dma.Release(uint(d.Address))
}

// Available represents a VirtIO virtual queue Available ring buffer.
//
// All exported fields are used one-time at initialization, fields requiring
// DMA are accessible through functions.
type Available struct {
	Flags      uint16
	index      uint16
	ring       []uint16
	EventIndex uint16

	// DMA buffer
	buf []byte
}

// Bytes converts the descriptor structure to byte array format.
func (d *Available) Bytes() []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, d.Flags)
	binary.Write(buf, binary.LittleEndian, d.index)

	for _, ring := range d.ring {
		binary.Write(buf, binary.LittleEndian, ring)
	}

	binary.Write(buf, binary.LittleEndian, d.EventIndex)

	return buf.Bytes()
}

// Index updates the descriptor index field.
func (d *Available) Index(index uint16) {
	off := 2
	binary.LittleEndian.PutUint16(d.buf[off:], index)

	d.index = index
}

// Ring returns a ring buffer at the given position.
func (d *Available) Ring(n uint16) uint16 {
	off := 4 + n*2
	d.ring[n] = binary.LittleEndian.Uint16(d.buf[off:])

	return d.ring[n]
}

// Set updates the index value of a ring buffer.
func (d *Available) Set(n uint16, index uint16) {
	off := 4 + n*2
	binary.LittleEndian.PutUint16(d.buf[off:], index)

	d.ring[n] = index
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

// Used represents a VirtIO virtual queue Used ring buffer.
//
// All exported fields are used one-time at initialization, fields requiring
// DMA are accessible through functions.
type Used struct {
	Flags      uint16
	index      uint16
	ring       []*Ring
	AvailEvent uint16

	// DMA buffer
	buf []byte

	last uint16
}

// Bytes converts the descriptor structure to byte array format.
func (d *Used) Bytes() []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, d.Flags)
	binary.Write(buf, binary.LittleEndian, d.index)

	for _, ring := range d.ring {
		buf.Write(ring.Bytes())
	}

	binary.Write(buf, binary.LittleEndian, d.AvailEvent)

	return buf.Bytes()
}

// Index returns the descriptor index field.
func (d *Used) Index() uint16 {
	off := 2
	d.index = binary.LittleEndian.Uint16(d.buf[off:])

	return d.index
}

// Ring returns a ring buffer at the given position.
func (d *Used) Ring(n uint16) Ring {
	off := 4 + n*8
	binary.Decode(d.buf[off:], binary.LittleEndian, d.ring[n])

	return *d.ring[n]
}

// VirtualQueue represents a VirtIO split virtual queue Descriptor
type VirtualQueue struct {
	sync.Mutex

	Descriptors []*Descriptor
	Available   Available
	Used        Used

	// DMA buffer
	buf    []byte
	desc   uint // physical address for QueueDesc
	driver uint // phusical address for QueueDriver
	device uint // physical address for QueueDevice

	size uint16
}

// Bytes converts the descriptor structure to byte array format, the device
// area and driver area location offsets are also returned.
func (d *VirtualQueue) Bytes() ([]byte, int, int) {
	buf := new(bytes.Buffer)

	for _, desc := range d.Descriptors {
		buf.Write(desc.Bytes())
	}

	driver := buf.Len()
	buf.Write(d.Available.Bytes())

	device := buf.Len()
	buf.Write(d.Used.Bytes())

	return buf.Bytes(), driver, device
}

// Init initializes a split virtual queue for the given size.
func (d *VirtualQueue) Init(size int, length int, flags uint16) {
	d.Lock()
	defer d.Unlock()

	for i := 0; i < size; i++ {
		desc := &Descriptor{}
		desc.Init(length, flags)

		ring := &Ring{}

		d.Descriptors = append(d.Descriptors, desc)
		d.Available.ring = append(d.Available.ring, uint16(i))
		d.Used.ring = append(d.Used.ring, ring)
	}

	if flags == Write {
		// make all buffers immediately available
		d.Available.index = uint16(size)
	}

	// allocate DMA buffer
	buf, driver, device := d.Bytes()
	d.desc, d.buf = dma.Reserve(len(buf), 16)
	copy(d.buf, buf)

	// calculate area pointers
	d.driver = d.desc + uint(driver)
	d.device = d.desc + uint(device)
	d.size = uint16(size)

	// assign DMA slices
	d.Available.buf = d.buf[driver:device]
	d.Used.buf = d.buf[device:]
}

// Destroy removes a split virtual queue from physical memory.
func (d *VirtualQueue) Destroy() {
	for _, d := range d.Descriptors {
		d.Destroy()
	}

	d.Available.buf = nil
	d.Used.buf = nil

	dma.Release(d.desc)
}

// Address returns the virtual queue physical address.
func (d *VirtualQueue) Address() (desc uint, driver uint, device uint) {
	return d.desc, d.driver, d.device
}

// Pop receives a single used buffer from the virtual queue,
func (d *VirtualQueue) Pop() (buf []byte) {
	d.Lock()
	defer d.Unlock()

	if d.Used.Index() == d.Used.last {
		return
	}

	avail := d.Used.Ring(d.Used.last % d.size)

	buf = make([]byte, avail.Length)
	copy(buf, d.Descriptors[avail.Index].buf)

	d.Available.index += 1
	d.Available.Set(d.Available.index%d.size, uint16(avail.Index))

	d.Available.Index(d.Available.index)
	d.Used.last += 1

	return
}

// Push supplies a single available buffer to the virtual queue.
func (d *VirtualQueue) Push(buf []byte) {
	d.Lock()
	defer d.Unlock()

	length := len(buf)
	index := d.Available.Ring(d.Available.index % d.size)

	d.Descriptors[index].Length(uint32(length))
	copy(d.Descriptors[index].buf, buf)

	d.Available.Index(d.Available.index + 1)

	for used := d.Used.Index() - d.Used.last; used > 0; used-- {
		index = used - 1
		avail := d.Used.Ring(used)

		d.Available.Set(d.Available.index%d.size, uint16(avail.Index))
		d.Used.last += 1
	}

	return
}

func (d *VirtualQueue) Debug() {
	fmt.Printf("\n%+v\n", d)

	for _, desc := range d.Descriptors {
		fmt.Printf("%x\n", desc)
	}

	for _, ring := range d.Used.ring {
		fmt.Printf("%x\n", ring)
	}

	descSize := len((&Descriptor{}).Bytes()) * len(d.Descriptors)
	availSize := len(d.Available.Bytes())

	driver := uint(descSize)
	device := driver + uint(availSize)

	fmt.Printf("%x", d.buf[device:])
}
