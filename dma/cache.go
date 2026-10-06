// First-fit memory allocator for DMA buffers
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !arm64

package dma

func cacheLineSize() uint {
	return 0
}

func (r *Region) clean(addr uint, size uint) {}

func (r *Region) invalidate(addr uint, size uint) {}
