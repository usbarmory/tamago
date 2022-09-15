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
