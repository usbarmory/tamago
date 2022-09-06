// First-fit memory allocator for DMA buffers
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package dma provides primitives for direct memory allocation and alignment,
// it is primarily used in bare metal device driver operation to avoid passing
// Go pointers for DMA purposes.
//
// This package is only meant to be used with `GOOS=tamago` as supported by the
// TamaGo framework for bare metal Go on ARM/RISC-V SoCs, see
// https://github.com/usbarmory/tamago.
package dma

import (
	"container/list"
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"unsafe"
)

type block struct {
	// pointer address
	addr uint
	// buffer size
	size uint
	// distinguish regular (`Alloc`/`Free`) and reserved
	// (`Reserve`/`Release`) blocks.
	res bool
}

// Region represents a memory region allocated for DMA purposes.
type Region struct {
	sync.Mutex

	start uint
	size  uint

	freeBlocks *list.List
	usedBlocks map[uint]*block
}

var dma *Region

// NewRegion initializes a memory region for DMA buffer allocation.
//
// To avoid unforseen consequences the caller must ensure that allocated
// regions do not overlap among themselves or with the global one (see Init()).
//
// To allow allocation of DMA buffers within Go runtime memory the unsafe flag
// must be set.
func NewRegion(addr uint32, size int, unsafe bool) (dma *Region, err error) {
	start := uint(addr)
	end := uint(start) + uint(size)

	// returns uint32/uint64 depending on platform
	rs, re := runtime.MemRegion()
	ramStart := uint(rs)
	ramEnd := uint(re)

	if !unsafe &&
		(ramStart > start && ramStart < end ||
			ramEnd > start && ramEnd < end ||
			start > ramStart && end < ramEnd) {
		return nil, fmt.Errorf("DMA within Go runtime memory (%#x-%#x) is not allowed", ramStart, ramEnd)
	}

	dma = &Region{
		start: start,
		size:  uint(size),
	}

	// initialize a single block to fit all available memory
	b := &block{
		addr: dma.start,
		size: dma.size,
	}

	dma.freeBlocks = list.New()
	dma.freeBlocks.PushFront(b)

	dma.usedBlocks = make(map[uint]*block)

	return
}

// Start returns the DMA region start address.
func (dma *Region) Start() uint32 {
	return uint32(dma.start)
}

// End returns the DMA region end address.
func (dma *Region) End() uint32 {
	return uint32(dma.start + dma.size)
}

// Size returns the DMA region size.
func (dma *Region) Size() uint32 {
	return uint32(dma.size)
}

// Reserve allocates a slice of bytes for DMA purposes, by placing its data
// within the DMA region, with optional alignment. It returns the slice along
// with its data allocation address. The buffer can be freed up with Release().
//
// Reserving buffers with Reserve() allows applications to pre-allocate DMA
// regions, avoiding unnecessary memory copy operations when performance is a
// concern. Reserved buffers cause Alloc() and Read() to return without any
// allocation or memory copy.
//
// Great care must be taken on reserved buffer as:
//   - buf contents are uninitialized (unlike when using Alloc())
//   - buf slices remain in reserved space but only the original buf
//     can be subject of Release()
//
// The optional alignment must be a power of 2 and word alignment is always
// enforced (0 == 4).
func (dma *Region) Reserve(size int, align int) (addr uint32, buf []byte) {
	if size == 0 {
		return
	}

	dma.Lock()
	defer dma.Unlock()

	b := dma.alloc(uint(size), uint(align))
	b.res = true

	dma.usedBlocks[b.addr] = b

	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	hdr.Data = uintptr(unsafe.Pointer(uintptr(b.addr)))
	hdr.Len = size
	hdr.Cap = hdr.Len

	return uint32(b.addr), buf
}

// Reserved returns whether a slice of bytes data is allocated within the DMA
// buffer region, it is used to determine whether the passed buffer has been
// previously allocated by this package with Reserve().
func (dma *Region) Reserved(buf []byte) (res bool, addr uint32) {
	ptr := uint(uintptr(unsafe.Pointer(&buf[0])))
	res = ptr >= dma.start && ptr+uint(len(buf)) <= dma.start+dma.size

	return res, uint32(ptr)
}

// Alloc reserves a memory region for DMA purposes, copying over a buffer and
// returning its allocation address, with optional alignment. The region can be
// freed up with Free().
//
// If the argument is a buffer previously created with Reserve(), then its
// address is return without any re-allocation.
//
// The optional alignment must be a power of 2 and word alignment is always
// enforced (0 == 4).
func (dma *Region) Alloc(buf []byte, align int) (addr uint32) {
	size := len(buf)

	if size == 0 {
		return 0
	}

	if res, addr := Reserved(buf); res {
		return uint32(addr)
	}

	dma.Lock()
	defer dma.Unlock()

	b := dma.alloc(uint(size), uint(align))
	b.write(0, buf)

	dma.usedBlocks[b.addr] = b

	return uint32(b.addr)
}

// Read reads exactly len(buf) bytes from a memory region address into a
// buffer, the region must have been previously allocated with Alloc().
//
// The offset and buffer size are used to retrieve a slice of the memory
// region, a panic occurs if these parameters are not compatible with the
// initial allocation for the address.
//
// If the argument is a buffer previously created with Reserve(), then the
// function returns without modifying it, as it is assumed for the buffer to be
// already updated.
func (dma *Region) Read(addr uint32, off int, buf []byte) {
	size := len(buf)

	if addr == 0 || size == 0 {
		return
	}

	if res, _ := Reserved(buf); res {
		return
	}

	dma.Lock()
	defer dma.Unlock()

	b, ok := dma.usedBlocks[uint(addr)]

	if !ok {
		panic("read of unallocated pointer")
	}

	if uint(off+size) > b.size {
		panic("invalid read parameters")
	}

	b.read(uint(off), buf)
}

// Write writes buffer contents to a memory region address, the region must
// have been previously allocated with Alloc().
//
// An offset can be passed to write a slice of the memory region, a panic
// occurs if the offset is not compatible with the initial allocation for the
// address.
func (dma *Region) Write(addr uint32, off int, buf []byte) {
	size := len(buf)

	if addr == 0 || size == 0 {
		return
	}

	dma.Lock()
	defer dma.Unlock()

	b, ok := dma.usedBlocks[uint(addr)]

	if !ok {
		return
	}

	if uint(off+size) > b.size {
		panic("invalid write parameters")
	}

	b.write(uint(off), buf)
}

// Free frees the memory region stored at the passed address, the region must
// have been previously allocated with Alloc().
func (dma *Region) Free(addr uint32) {
	dma.freeBlock(uint(addr), false)
}

// Release frees the memory region stored at the passed address, the region
// must have been previously allocated with Reserve().
func (dma *Region) Release(addr uint32) {
	dma.freeBlock(uint(addr), true)
}

func (dma *Region) freeBlock(addr uint, res bool) {
	if addr == 0 {
		return
	}

	dma.Lock()
	defer dma.Unlock()

	b, ok := dma.usedBlocks[addr]

	if !ok {
		return
	}

	if b.res != res {
		return
	}

	dma.free(b)
	delete(dma.usedBlocks, addr)
}

// Init initializes the global memory region for DMA buffer allocation, used
// throughout the tamago package for all DMA allocations.
//
// Additional DMA regions for application use can be allocated through
// NewRegion().
func Init(start uint32, size int) (err error) {
	dma, err = NewRegion(start, size, false)
	return
}

// Default returns the global DMA region instance.
func Default() *Region {
	return dma
}

// Reserve is the equivalent of Region.Reserve() on the global DMA region.
func Reserve(size int, align int) (addr uint32, buf []byte) {
	return dma.Reserve(size, align)
}

// Reserved is the equivalent of Region.Reserved() on the global DMA region.
func Reserved(buf []byte) (res bool, addr uint32) {
	return dma.Reserved(buf)
}

// Alloc is the equivalent of Region.Alloc() on the global DMA region.
func Alloc(buf []byte, align int) (addr uint32) {
	return dma.Alloc(buf, align)
}

// Read is the equivalent of Region.Read() on the global DMA region.
func Read(addr uint32, off int, buf []byte) {
	dma.Read(addr, off, buf)
}

// Write is the equivalent of Region.Write() on the global DMA region.
func Write(addr uint32, off int, buf []byte) {
	dma.Write(addr, off, buf)
}

// Free is the equivalent of Region.Free() on the global DMA region.
func Free(addr uint32) {
	dma.Free(addr)
}

// Release is the equivalent of Region.Release() on the global DMA region.
func Release(addr uint32) {
	dma.Release(addr)
}
