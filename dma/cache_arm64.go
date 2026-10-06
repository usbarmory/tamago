// First-fit memory allocator for DMA buffers
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package dma

// defined in cache_arm64.s
func cache_line_size() uint64
func cache_clean(start uint64, end uint64, line uint64)
func cache_invalidate(start uint64, end uint64, line uint64)

// A region may be mapped as cacheable memory (see arm64.CPU.ConfigureMMU)
// while devices which are not cache coherent access it, therefore buffers are
// aligned to the data cache line size, written back before devices read them
// and discarded before the CPU reads what devices wrote.

func cacheLineSize() uint {
	return uint(cache_line_size())
}

func (r *Region) clean(addr uint, size uint) {
	if size == 0 {
		return
	}

	cache_clean(uint64(addr&^(r.line-1)), uint64(addr+size), uint64(r.line))
}

func (r *Region) invalidate(addr uint, size uint) {
	if size == 0 {
		return
	}

	cache_invalidate(uint64(addr&^(r.line-1)), uint64(addr+size), uint64(r.line))
}
