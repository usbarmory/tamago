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
	"sync"
	"unsafe"
)

// Block represents a memory I/O allocation block.
type Block interface {
	// Address returns the block address.
	Address() uint

	// Size returns the block length.
	Size() uint

	// Grow increases the block length by the size argument.
	Grow(size uint)

	// Shrink reduces the block length by the size argument shifting its
	// address by the specified offset.
	Shrink(off uint, size uint)

	// Read reads up to len(buf) into buf from the block address offset.
	Read(off uint, buf []byte)

	// Write writes len(buf) bytes from buf to the block address offset.
	Write(off uint, buf []byte)

	// Slice returns a slice whose underlying array starts at the block
	// pointer and whose length and capacity match the block size.
	Slice() []byte
}

// Region represents a memory region allocated for DMA purposes.
type Region struct {
	sync.Mutex

	Block func(addr uint, size uint) Block

	start uint
	size  uint

	freeBlocks *list.List
	usedBlocks map[uint]Block
}

var dma *Region

// Default returns the global DMA region instance.
func Default() *Region {
	return dma
}

// Init initializes a memory region with a single block to fit all available
// memory.
func (r *Region) Init(start uint, size uint) {
	r.start = start
	r.size = size

	b := r.Block(start, size)

	r.freeBlocks = list.New()
	r.freeBlocks.PushFront(b)

	r.usedBlocks = make(map[uint]Block)
}

// Start returns the DMA region start address.
func (r *Region) Start() uint {
	return r.start
}

// End returns the DMA region end address.
func (r *Region) End() uint {
	return r.start + r.size
}

// Size returns the DMA region size.
func (r *Region) Size() uint {
	return r.size
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
func (r *Region) Reserve(size int, align int) (addr uint, buf []byte) {
	if size == 0 {
		return
	}

	r.Lock()
	defer r.Unlock()

	b := r.alloc(uint(size), uint(align))
	addr = b.Address()

	r.usedBlocks[addr] = b

	return addr, b.Slice()
}

// Reserved returns whether a slice of bytes data is allocated within the DMA
// buffer region, it is used to determine whether the passed buffer has been
// previously allocated by this package with Reserve().
func (r *Region) Reserved(buf []byte) (res bool, addr uint) {
	addr = uint(uintptr(unsafe.Pointer(&buf[0])))
	res = addr >= r.start && addr+uint(len(buf)) <= r.start+r.size

	return
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
func (r *Region) Alloc(buf []byte, align int) (addr uint) {
	size := len(buf)

	if size == 0 {
		return 0
	}

	if res, addr := Reserved(buf); res {
		return addr
	}

	r.Lock()
	defer r.Unlock()

	b := r.alloc(uint(size), uint(align))
	b.Write(0, buf)
	addr = b.Address()

	r.usedBlocks[addr] = b

	return
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
func (r *Region) Read(addr uint, off int, buf []byte) {
	size := len(buf)

	if addr == 0 || size == 0 {
		return
	}

	if res, _ := Reserved(buf); res {
		return
	}

	r.Lock()
	defer r.Unlock()

	b, ok := r.usedBlocks[addr]

	if !ok {
		panic("read of unallocated pointer")
	}

	if uint(off+size) > b.Size() {
		panic("invalid read parameters")
	}

	b.Read(uint(off), buf)
}

// Write writes buffer contents to a memory region address, the region must
// have been previously allocated with Alloc().
//
// An offset can be passed to write a slice of the memory region, a panic
// occurs if the offset is not compatible with the initial allocation for the
// address.
func (r *Region) Write(addr uint, off int, buf []byte) {
	size := len(buf)

	if addr == 0 || size == 0 {
		return
	}

	r.Lock()
	defer r.Unlock()

	b, ok := r.usedBlocks[addr]

	if !ok {
		return
	}

	if uint(off+size) > b.Size() {
		panic("invalid write parameters")
	}

	b.Write(uint(off), buf)
}

// Free frees the memory region stored at the passed address, the region must
// have been previously allocated with Alloc().
func (r *Region) Free(addr uint) {
	r.freeBlock(addr)
}

// Release frees the memory region stored at the passed address, the region
// must have been previously allocated with Reserve().
func (r *Region) Release(addr uint) {
	r.freeBlock(addr)
}

func (r *Region) defrag() {
	var prevBlock Block

	// find contiguous free blocks and combine them
	for e := r.freeBlocks.Front(); e != nil; e = e.Next() {
		b := e.Value.(Block)

		if prevBlock != nil {
			if prevBlock.Address()+prevBlock.Size() == b.Address() {
				prevBlock.Grow(b.Size())
				defer r.freeBlocks.Remove(e)
				continue
			}
		}

		prevBlock = e.Value.(Block)
	}
}

func (r *Region) alloc(size uint, align uint) Block {
	var e *list.Element
	var freeBlock Block
	var pad uint

	if align == 0 {
		// force word alignment
		align = 4
	}

	// find suitable block
	for e = r.freeBlocks.Front(); e != nil; e = e.Next() {
		b := e.Value.(Block)

		// pad to required alignment
		pad = -b.Address() & (align - 1)

		if b.Size() >= size+pad {
			freeBlock = b
			size += pad
			break
		}
	}

	if freeBlock == nil {
		panic("out of memory")
	}

	// allocate block from free linked list
	defer r.freeBlocks.Remove(e)

	// adjust block to desired size, add new block for remainder
	if n := freeBlock.Size() - size; n != 0 {
		newBlockAfter := r.Block(freeBlock.Address()+size, n)

		freeBlock.Shrink(0, n)
		r.freeBlocks.InsertAfter(newBlockAfter, e)
	}

	if pad != 0 {
		// claim padding space
		newBlockBefore := r.Block(freeBlock.Address(), pad)

		freeBlock.Shrink(pad, pad)
		r.freeBlocks.InsertBefore(newBlockBefore, e)
	}

	return freeBlock
}

func (r *Region) free(usedBlock Block) {
	for e := r.freeBlocks.Front(); e != nil; e = e.Next() {
		b := e.Value.(Block)

		if b.Address() > usedBlock.Address() {
			r.freeBlocks.InsertBefore(usedBlock, e)
			r.defrag()
			return
		}
	}

	r.freeBlocks.PushBack(usedBlock)
}

func (r *Region) freeBlock(addr uint) {
	if addr == 0 {
		return
	}

	r.Lock()
	defer r.Unlock()

	b, ok := r.usedBlocks[addr]

	if !ok {
		return
	}

	r.free(b)
	delete(r.usedBlocks, addr)
}
