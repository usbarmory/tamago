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
)

// Init initializes a memory region for DMA buffer allocation, the application
// must guarantee that the passed memory range is never used by the Go
// runtime (defining `runtime.ramStart` and `runtime.ramSize` accordingly).
func Init(start uint32, size int) {
	mutex.Lock()
	// note: cannot defer during initialization

	// initialize a single block to fit all available memory
	b := &block{
		addr: start,
		size: size,
	}

	freeBlocks = list.New()
	freeBlocks.PushFront(b)

	usedBlocks = make(map[uint32]*block)

	defer mutex.Unlock()
}

// Alloc reserves a memory region, copies over a buffer and return its
// allocation address, with optional alignment. The region can be freed up with
// `Free`.
func Alloc(buf []byte, align int) uint32 {
	mutex.Lock()
	defer mutex.Unlock()

	size := len(buf)

	if size == 0 {
		return 0
	}

	b := alloc(len(buf), align)
	b.write(buf, 0)

	usedBlocks[b.addr] = b

	return b.addr
}

// Read reads exactly len(buf) bytes from a memory region address into buf, the
// region must have been previously allocated with `Alloc`.
//
// The offset and buffer size are used to retrieve a slice of the memory
// region, a panic occurs if these parameters are not compatible with the
// initial allocation for the address.
func Read(addr uint32, offset int, buf []byte) {
	mutex.Lock()
	defer mutex.Unlock()

	if addr == 0 {
		return
	}

	b, ok := usedBlocks[addr]

	if !ok {
		panic("read of unallocated pointer")
	}

	if offset+len(buf) > b.size {
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
	mutex.Lock()
	defer mutex.Unlock()

	size := len(data)

	if addr == 0 || size == 0 {
		return
	}

	b, ok := usedBlocks[addr]

	if !ok {
		panic("write of unallocated pointer")
	}

	if offset+size > b.size {
		panic("invalid write parameters")
	}

	b.write(data, offset)
}

// Free frees the memory region stored at the passed address, the region must
// have been previously allocated with `Alloc`. A region can only be freed
// once, otherwise a panic occurs.
func Free(addr uint32) {
	mutex.Lock()
	defer mutex.Unlock()

	if addr == 0 {
		return
	}

	b, ok := usedBlocks[addr]

	if !ok {
		panic("free of unallocated pointer")
	}

	free(b)
	delete(usedBlocks, addr)
}
