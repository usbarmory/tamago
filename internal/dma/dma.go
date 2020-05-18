// First-fit memory allocator for DMA buffer allocation
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
	addr uint32
	size int
}

var freeBlocks *list.List
var usedBlocks map[uint32]*block

var mutex sync.Mutex

func (b *block) read(offset int, buf []byte) {
	var mem []byte

	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&mem))
	hdr.Data = uintptr(unsafe.Pointer(uintptr(b.addr + uint32(offset))))
	hdr.Len = len(buf)
	hdr.Cap = hdr.Len

	copy(buf, mem)
}

func (b *block) write(buf []byte, offset int) {
	var mem []byte

	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&mem))
	hdr.Data = uintptr(unsafe.Pointer(uintptr(b.addr + uint32(offset))))
	hdr.Len = len(buf)

	copy(mem, buf)
}

func defrag() {
	var prevBlock *block

	// find contiguous free blocks and combine them
	for e := freeBlocks.Front(); e != nil; e = e.Next() {
		b := e.Value.(*block)

		if prevBlock != nil {
			if prevBlock.addr+uint32(prevBlock.size) == b.addr {
				prevBlock.size += b.size
				defer freeBlocks.Remove(e)
				continue
			}
		}

		prevBlock = e.Value.(*block)
	}
}

func alloc(size int, align int) *block {
	var e *list.Element
	var freeBlock *block

	// make room for alignment buffer
	if align > 0 {
		size += align
	}

	// find suitable block
	for e = freeBlocks.Front(); e != nil; e = e.Next() {
		b := e.Value.(*block)

		if b.size >= size {
			freeBlock = b
			break
		}
	}

	if freeBlock == nil {
		panic("out of memory")
	}

	// when we are done remove block from free linked list
	defer freeBlocks.Remove(e)

	// adjust block to desired size, add new block to leave remainder
	if size < freeBlock.size {
		newBlockAfter := &block{
			addr: freeBlock.addr + uint32(size),
			size: freeBlock.size - size,
		}

		freeBlock.size = size
		freeBlocks.InsertAfter(newBlockAfter, e)
	}

	if align > 0 {
		if r := int(freeBlock.addr) & (align - 1); r != 0 {
			offset := align - r

			// claim space between block address and alignment offset
			newBlockBefore := &block{
				addr: freeBlock.addr,
				size: offset,
			}

			freeBlock.addr += uint32(offset)
			freeBlock.size -= offset
			freeBlocks.InsertBefore(newBlockBefore, e)
		}

		// original requested size
		size -= align

		// claim back leftover from alignment buffer
		if freeBlock.size > size {
			newBlockAfter := &block{
				addr: freeBlock.addr + uint32(size),
				size: freeBlock.size - size,
			}

			freeBlock.size = size
			freeBlocks.InsertAfter(newBlockAfter, e)
		}
	}

	return freeBlock
}

func free(usedBlock *block) {
	for e := freeBlocks.Front(); e != nil; e = e.Next() {
		b := e.Value.(*block)

		if b.addr > usedBlock.addr {
			freeBlocks.InsertBefore(usedBlock, e)
			defrag()
			return
		}
	}

	freeBlocks.PushBack(usedBlock)
}

// Init initializes a memory region for DMA buffer allocation.
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
