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
	"reflect"
	"unsafe"
)

type memoryBlock struct {
	// pointer address
	addr uint
	// buffer size
	size uint
	// distinguish regular (`Alloc`/`Free`) and reserved
	// (`Reserve`/`Release`) blocks.
	res bool
}

func newMemoryBlock(addr uint, size uint) Block {
	return &memoryBlock{
		addr: addr,
		size: size,
	}
}

func (b *memoryBlock) Address() uint {
	return b.addr
}

func (b *memoryBlock) Size() uint {
	return b.size
}

func (b *memoryBlock) Grow(size uint) {
	b.size += size
}

func (b *memoryBlock) Shrink(off uint, size uint) {
	b.addr += off
	b.size -= size
}

func (b *memoryBlock) Read(off uint, buf []byte) {
	var ptr unsafe.Pointer

	ptr = unsafe.Add(ptr, b.addr+off)
	mem := unsafe.Slice((*byte)(ptr), len(buf))

	copy(buf, mem)
}

func (b *memoryBlock) Write(off uint, buf []byte) {
	var ptr unsafe.Pointer

	ptr = unsafe.Add(ptr, b.addr+off)
	mem := unsafe.Slice((*byte)(ptr), len(buf))

	copy(mem, buf)
}

func (b *memoryBlock) Slice() (buf []byte) {
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	hdr.Data = uintptr(unsafe.Pointer(uintptr(b.addr)))
	hdr.Len = int(b.size)
	hdr.Cap = hdr.Len

	return
}
