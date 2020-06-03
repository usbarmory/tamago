// First-fit memory allocator for DMA buffers
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package dma

import (
	"container/list"
	"reflect"
	"unsafe"
)

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
	for e := dma.freeBlocks.Front(); e != nil; e = e.Next() {
		b := e.Value.(*block)

		if prevBlock != nil {
			if prevBlock.addr+uint32(prevBlock.size) == b.addr {
				prevBlock.size += b.size
				defer dma.freeBlocks.Remove(e)
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
	for e = dma.freeBlocks.Front(); e != nil; e = e.Next() {
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
	defer dma.freeBlocks.Remove(e)

	// adjust block to desired size, add new block to leave remainder
	if size < freeBlock.size {
		newBlockAfter := &block{
			addr: freeBlock.addr + uint32(size),
			size: freeBlock.size - size,
		}

		freeBlock.size = size
		dma.freeBlocks.InsertAfter(newBlockAfter, e)
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
			dma.freeBlocks.InsertBefore(newBlockBefore, e)
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
			dma.freeBlocks.InsertAfter(newBlockAfter, e)
		}
	}

	return freeBlock
}

func free(usedBlock *block) {
	for e := dma.freeBlocks.Front(); e != nil; e = e.Next() {
		b := e.Value.(*block)

		if b.addr > usedBlock.addr {
			dma.freeBlocks.InsertBefore(usedBlock, e)
			defrag()
			return
		}
	}

	dma.freeBlocks.PushBack(usedBlock)
}
