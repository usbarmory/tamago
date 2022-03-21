// First-fit memory allocator for DMA buffers
// https://github.com/usbarmory/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package dma

import (
	"container/list"
	"unsafe"
)

func (b *block) read(off int, buf []byte) {
	var ptr unsafe.Pointer

	ptr = unsafe.Add(ptr, int(b.addr) + off)
	mem := unsafe.Slice((*byte)(ptr), len(buf))

	copy(buf, mem)
}

func (b *block) write(off int, buf []byte) {
	var ptr unsafe.Pointer

	ptr = unsafe.Add(ptr, int(b.addr) + off)
	mem := unsafe.Slice((*byte)(ptr), len(buf))

	copy(mem, buf)
}

func (dma *Region) defrag() {
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

func (dma *Region) alloc(size int, align int) *block {
	var e *list.Element
	var freeBlock *block
	var pad int

	if align == 0 {
		// force word alignment
		align = 4
	}

	// find suitable block
	for e = dma.freeBlocks.Front(); e != nil; e = e.Next() {
		b := e.Value.(*block)

		// pad to required alignment
		pad = -int(b.addr) & (align - 1)
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
			addr: freeBlock.addr + uint32(size),
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

		freeBlock.addr += uint32(pad)
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
