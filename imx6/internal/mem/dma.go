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

// Package mem provides primitives for direct memory allocation and alignment,
// it is primarily used in bare metal device driver operation to avoid passing
// Go pointers for DMA purposes.
package mem

import (
	"container/list"
	"sync"
	"unsafe"
)

const iramStart uint32 = 0x00900000
const iramSize = 0x20000

type block struct {
	addr uint32
	size int
}

var freeBlocks *list.List
var usedBlocks map[uint32]*block

var mutex sync.Mutex

func (b *block) read(offset int, size int) []byte {
	data := make([]byte, size)

	for i := 0; i < size; i++ {
		data[i] = *(*byte)(unsafe.Pointer(uintptr(b.addr + uint32(offset+i))))
	}

	return data
}

func (b *block) write(data []byte, offset int) {
	for i := range data {
		*(*byte)(unsafe.Pointer(uintptr(b.addr + uint32(offset+i)))) = data[i]
	}
}

func init() {
	// use internal OCRAM (iRAM) by default
	Init(iramStart, iramSize)
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

func check(addr int, align int) bool {
	return addr&(align-1) == 0
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

// Alloc reserves a memory region, copies over a data buffer and return its
// allocation address, with optional alignment. The region can be freed up with
// `Free`.
func Alloc(data []byte, align int) uint32 {
	mutex.Lock()
	defer mutex.Unlock()

	size := len(data)

	if size == 0 {
		return 0
	}

	b := alloc(len(data), align)
	b.write(data, 0)

	usedBlocks[b.addr] = b

	return b.addr
}

// Read returns the data buffer stored at the corresponding memory region
// address, the region must have been previously allocated with `Alloc`.
//
// The offset and size are used to retrieve a slice of the buffer, a panic
// occurs if these parameters are not compatible with the initial allocation
// for the address.
func Read(addr uint32, offset int, size int) []byte {
	mutex.Lock()
	defer mutex.Unlock()

	if addr == 0 {
		return []byte{}
	}

	b, ok := usedBlocks[addr]

	if !ok {
		panic("read of unallocated pointer")
	}

	if offset+size > b.size {
		panic("invalid read parameters")
	}

	return b.read(offset, size)
}

// Write writes in the data buffer stored at the corresponding memory region
// address, the region must have been previously allocated with `Alloc`.
//
// An offset can be pased to write a slice of the buffer, a panic occurs if the
// offset is not compatible with the initial allocation for the address.
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
