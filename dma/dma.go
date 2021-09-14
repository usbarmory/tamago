// First-fit memory allocator for DMA buffers
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package dma provides primitives for direct memory allocation and alignment,
// it is primarily used in bare metal device driver operation to avoid passing
// Go pointers for DMA purposes.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/f-secure-foundry/tamago.
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

// Region represents a memory region allocated for DMA purposes.
type Region struct {
	sync.Mutex

	Start uint32
	Size  int

	freeBlocks *list.List
	usedBlocks map[uint32]*block
}

var dma *Region

// Init initializes a memory region for DMA buffer allocation, the application
// must guarantee that the passed memory range is never used by the Go
// runtime (defining runtime.ramStart and runtime.ramSize accordingly).
func (dma *Region) Init() {
	// initialize a single block to fit all available memory
	b := &block{
		addr: dma.Start,
		size: dma.Size,
	}

	dma.Lock()
	defer dma.Unlock()

	dma.freeBlocks = list.New()
	dma.freeBlocks.PushFront(b)

	dma.usedBlocks = make(map[uint32]*block)
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
//   * buf contents are uninitialized (unlike when using Alloc())
//   * buf slices remain in reserved space but only the original buf
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

	b := dma.alloc(size, align)
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
func (dma *Region) Reserved(buf []byte) (res bool, addr uint32) {
	addr = uint32(uintptr(unsafe.Pointer(&buf[0])))
	res = addr >= dma.Start && addr+uint32(len(buf)) <= dma.Start+uint32(dma.Size)

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
func (dma *Region) Alloc(buf []byte, align int) (addr uint32) {
	size := len(buf)

	if size == 0 {
		return 0
	}

	if res, addr := Reserved(buf); res {
		return addr
	}

	dma.Lock()
	defer dma.Unlock()

	b := dma.alloc(len(buf), align)
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

	b, ok := dma.usedBlocks[addr]

	if !ok {
		panic("read of unallocated pointer")
	}

	if off+size > b.size {
		panic("invalid read parameters")
	}

	b.read(off, buf)
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

	b, ok := dma.usedBlocks[addr]

	if !ok {
		return
	}

	if off+size > b.size {
		panic("invalid write parameters")
	}

	b.write(off, buf)
}

// Free frees the memory region stored at the passed address, the region must
// have been previously allocated with Alloc().
func (dma *Region) Free(addr uint32) {
	dma.freeBlock(addr, false)
}

// Release frees the memory region stored at the passed address, the region
// must have been previously allocated with Reserve().
func (dma *Region) Release(addr uint32) {
	dma.freeBlock(addr, true)
}

func (dma *Region) freeBlock(addr uint32, res bool) {
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

// Init initializes the global memory region for DMA buffer allocation, the
// application must guarantee that the passed memory range is never used by the
// Go runtime (defining runtime.ramStart and runtime.ramSize accordingly).
//
// The global region is used throughout the tamago package for all DMA
// allocations.
//
// Separate DMA regions can be allocated in other areas (e.g. external RAM) by
// the application using Region.Init().
func Init(start uint32, size int) {
	dma = &Region{
		Start: start,
		Size:  size,
	}

	dma.Init()
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
