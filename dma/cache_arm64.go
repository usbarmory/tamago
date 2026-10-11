// First-fit memory allocator for DMA buffers
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package dma

import (
	"github.com/usbarmory/tamago/internal/cache"
)

func cacheGranule() uint {
	return uint(cache.WritebackGranule())
}

func (r *Region) clean(addr uint, size uint) {
	if r.granule == 0 {
		return
	}

	cache.Clean(uint64(addr), int(size))
}

func (r *Region) invalidate(addr uint, size uint) {
	if r.granule == 0 {
		return
	}

	cache.Invalidate(uint64(addr), int(size))
}
