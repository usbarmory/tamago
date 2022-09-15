// First-fit memory allocator for DMA buffers
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package dma

import (
	"container/list"
	"reflect"
	"sync"
	"unsafe"
)

// Region represents a memory region allocated for DMA purposes.
type Region struct {
	sync.Mutex

	start uint
	size  uint

	freeBlocks *list.List
	usedBlocks map[uint]*block
}

var dma *Region

// Default returns the global DMA region instance.
func Default() *Region {
	return dma
}

// Start returns the DMA region start address.
func (dma *Region) Start() uint {
	return dma.start
}

// End returns the DMA region end address.
func (dma *Region) End() uint {
	return dma.start + dma.size
}

// Size returns the DMA region size.
func (dma *Region) Size() uint {
	return dma.size
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
func (dma *Region) Reserve(size int, align int) (addr uint, buf []byte) {
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

	return b.addr, buf
}

// Reserved returns whether a slice of bytes data is allocated within the DMA
// buffer region, it is used to determine whether the passed buffer has been
// previously allocated by this package with Reserve().
func (dma *Region) Reserved(buf []byte) (res bool, addr uint) {
	ptr := uint(uintptr(unsafe.Pointer(&buf[0])))
	res = ptr >= dma.start && ptr+uint(len(buf)) <= dma.start+dma.size

	return res, ptr
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
func (dma *Region) Alloc(buf []byte, align int) (addr uint) {
	size := len(buf)

	if size == 0 {
		return 0
	}

	if res, addr := Reserved(buf); res {
		return addr
	}

	dma.Lock()
	defer dma.Unlock()

	b := dma.alloc(uint(size), uint(align))
	b.write(0, buf)

	dma.usedBlocks[b.addr] = b

	return b.addr
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
func (dma *Region) Read(addr uint, off int, buf []byte) {
	size := len(buf)

	if addr == 0 || size == 0 {
		return
	}

	if res, _ := Reserved(buf); res {
		return
	}

	dma.Lock()
	defer dma.Unlock()

	b, ok := dma.usedBlocks[addr]

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
func (dma *Region) Write(addr uint, off int, buf []byte) {
	size := len(buf)

	if addr == 0 || size == 0 {
		return
	}

	dma.Lock()
	defer dma.Unlock()

	b, ok := dma.usedBlocks[addr]

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
func (dma *Region) Free(addr uint) {
	dma.freeBlock(addr, false)
}

// Release frees the memory region stored at the passed address, the region
// must have been previously allocated with Reserve().
func (dma *Region) Release(addr uint) {
	dma.freeBlock(addr, true)
}

func (dma *Region) defrag() {
	var prevBlock *block

	// find contiguous free blocks and combine them
	for e := dma.freeBlocks.Front(); e != nil; e = e.Next() {
		b := e.Value.(*block)

		if prevBlock != nil {
			if prevBlock.addr+prevBlock.size == b.addr {
				prevBlock.size += b.size
				defer dma.freeBlocks.Remove(e)
				continue
			}
		}

		prevBlock = e.Value.(*block)
	}
}

func (dma *Region) alloc(size uint, align uint) *block {
	var e *list.Element
	var freeBlock *block
	var pad uint

	if align == 0 {
		// force word alignment
		align = 4
	}

	// find suitable block
	for e = dma.freeBlocks.Front(); e != nil; e = e.Next() {
		b := e.Value.(*block)

		// pad to required alignment
		pad = -b.addr & (align - 1)
		size += pad

		if b.size >= size {
			freeBlock = b
			break
		}
	}

	if freeBlock == nil {
		panic("out of memory")
	}

	// allocate block from free linked list
	defer dma.freeBlocks.Remove(e)

	// adjust block to desired size, add new block for remainder
	if r := freeBlock.size - size; r != 0 {
		newBlockAfter := &block{
			addr: freeBlock.addr + size,
			size: r,
		}

		freeBlock.size = size
		dma.freeBlocks.InsertAfter(newBlockAfter, e)
	}

	if pad != 0 {
		// claim padding space
		newBlockBefore := &block{
			addr: freeBlock.addr,
			size: pad,
		}

		freeBlock.addr += pad
		freeBlock.size -= pad
		dma.freeBlocks.InsertBefore(newBlockBefore, e)
	}

	return freeBlock
}

func (dma *Region) free(usedBlock *block) {
	for e := dma.freeBlocks.Front(); e != nil; e = e.Next() {
		b := e.Value.(*block)

		if b.addr > usedBlock.addr {
			dma.freeBlocks.InsertBefore(usedBlock, e)
			dma.defrag()
			return
		}
	}

	dma.freeBlocks.PushBack(usedBlock)
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
