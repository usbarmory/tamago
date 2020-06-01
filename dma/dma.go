// First-fit memory allocator for DMA buffers
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build tamago,arm

// Package dma provides primitives for direct memory allocation and alignment,
// it is primarily used in bare metal device driver operation to avoid passing
// Go pointers for DMA purposes.
package dma

import (
	"container/list"
	"reflect"
	"sync"
	"unsafe"
)

type block struct {
	// pointer address
	addr uint32
	// buffer size
	size int
	// distinguish regular (`Alloc`/`Free`) and reserved
	// (`Reserve`/`Release`) blocks.
	res bool
}

type region struct {
	sync.Mutex

	start uint32
	size  int

	freeBlocks *list.List
	usedBlocks map[uint32]*block
}

var dma = &region{}

// Init initializes a memory region for DMA buffer allocation, the application
// must guarantee that the passed memory range is never used by the Go
// runtime (defining `runtime.ramStart` and `runtime.ramSize` accordingly).
func Init(start uint32, size int) {
	dma.Lock()
	// note: cannot defer during initialization

	dma.start = start
	dma.size = size

	// initialize a single block to fit all available memory
	b := &block{
		addr: start,
		size: size,
	}

	dma.freeBlocks = list.New()
	dma.freeBlocks.PushFront(b)

	dma.usedBlocks = make(map[uint32]*block)

	dma.Unlock()
}

// Reserve allocated a slice of bytes for DMA purposes, by placing its data
// within the DMA region, with optional alignment. It returns the slice along
// with its data allocation address. The buffer can be freed up with `Release`.
//
// Reserving buffers with `Reserve` allows applications to pre-allocate DMA
// regions, avoiding unnecessary memory copy operations when performance is a
// concern. Reserved buffers cause `Alloc` and `Read` to return without any
// allocation or memory copy.
func Reserve(size int, align int) (addr uint32, buf []byte) {
	dma.Lock()
	defer dma.Unlock()

	if size == 0 {
		return
	}

	b := alloc(size, align)
	b.res = true

	dma.usedBlocks[b.addr] = b

	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	hdr.Data = uintptr(unsafe.Pointer(uintptr(b.addr)))
	hdr.Len = size
	hdr.Cap = hdr.Len

	return b.addr, buf
}

// Reserved returns whether a slice of bytes data is allocated within the DMA
// buffer region, it is used to determine whether the passed buffer has been
// previously allocated by this package with `Reserve`.
func Reserved(buf []byte) (res bool, addr uint32) {
	addr = uint32(uintptr(unsafe.Pointer(&buf[0])))
	res = addr >= dma.start && addr+uint32(len(buf)) <= dma.start+uint32(dma.size)

	return
}

// Alloc reserves a memory region for DMA purposes, copying over a buffer and
// returning its allocation address, with optional alignment. The region can be
// freed up with `Free`.
//
// If the argument is a buffer previously created with `Reserve`, then
// its address is return without any re-allocation.
func Alloc(buf []byte, align int) (addr uint32) {
	dma.Lock()
	defer dma.Unlock()

	size := len(buf)

	if size == 0 {
		return 0
	}

	if res, addr := Reserved(buf); res {
		return addr
	}

	b := alloc(len(buf), align)
	b.write(buf, 0)

	dma.usedBlocks[b.addr] = b

	return b.addr
}

// Read reads exactly len(buf) bytes from a memory region address into a
// buffer, the region must have been previously allocated with `Alloc`.
//
// The offset and buffer size are used to retrieve a slice of the memory
// region, a panic occurs if these parameters are not compatible with the
// initial allocation for the address.
//
// If the argument is a buffer previously created with `Reserve`, then the
// function returns without modifying it, as it is assumed for the buffer to be
// already updated.
func Read(addr uint32, offset int, buf []byte) {
	dma.Lock()
	defer dma.Unlock()

	size := len(buf)

	if addr == 0 || size == 0 {
		return
	}

	if res, _ := Reserved(buf); res {
		return
	}

	b, ok := dma.usedBlocks[addr]

	if !ok {
		panic("read of unallocated pointer")
	}

	if offset+size > b.size {
		panic("invalid read parameters")
	}

	b.read(offset, buf)
}

// Write writes buffer contents to a memory region address, the region must
// have been previously allocated with `Alloc`.
//
// An offset can be pased to write a slice of the memory region, a panic occurs
// if the offset is not compatible with the initial allocation for the address.
func Write(addr uint32, data []byte, offset int) {
	dma.Lock()
	defer dma.Unlock()

	size := len(data)

	if addr == 0 || size == 0 {
		return
	}

	b, ok := dma.usedBlocks[addr]

	if !ok {
		return
	}

	if offset+size > b.size {
		panic("invalid write parameters")
	}

	b.write(data, offset)
}

// Free frees the memory region stored at the passed address, the region must
// have been previously allocated with `Alloc`.
func Free(addr uint32) {
	freeBlock(addr, false)
}

// Release frees the memory region stored at the passed address, the region
// must have been previously allocated with `Reserve`.
func Release(addr uint32) {
	freeBlock(addr, true)
}

func freeBlock(addr uint32, res bool) {
	dma.Lock()
	defer dma.Unlock()

	if addr == 0 {
		return
	}

	b, ok := dma.usedBlocks[addr]

	if !ok {
		return
	}

	if b.res != res {
		return
	}

	free(b)
	delete(dma.usedBlocks, addr)
}
