// First-fit memory allocator for DMA buffers
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package dma

import "unsafe"

const chunk = 960

type block struct {
	// pointer address
	addr uint
	// buffer size
	size uint
	// distinguish regular (`Alloc`/`Free`) and reserved
	// (`Reserve`/`Release`) blocks.
	res bool
}

// On arm64 Go `copy` built-in cannot be safely used on device memory
// as LDP/STP instructions require 8-byte alignment, for this reason
// all `copy` against DMA buffers are forced to 8-byte aligned slices.
//
// From 1 KiB the arm64 `copy` also aligns its destination and moves the
// source by the same amount, so slices are copied in smaller chunks.
func (b *block) copy(dst, src []byte, align uint, off uint, write bool) {
	var split bool

	n := len(src)
	r := n % int(align)
	n -= r

	switch {
	case off%align != 0:
		split = true
	case write:
		if addr := uint(uintptr(unsafe.Pointer(&src[0]))); addr%align != 0 {
			split = true
		}
	default:
		if addr := uint(uintptr(unsafe.Pointer(&dst[0]))); addr%align != 0 {
			split = true
		}
	}

	if split {
		for i := 0; i < n; i += chunk {
			end := min(i+chunk, n)
			copy(dst[i:end], src[i:end])
		}
	} else {
		copy(dst, src[:n])
	}

	for i := range r {
		dst[n+i] = src[n+i]
	}
}

func (b *block) read(off uint, buf []byte) {
	var ptr unsafe.Pointer

	ptr = unsafe.Add(ptr, b.addr+off)
	mem := unsafe.Slice((*byte)(ptr), len(buf))

	b.copy(buf, mem, 8, off, false)
}

func (b *block) write(off uint, buf []byte) {
	var ptr unsafe.Pointer

	ptr = unsafe.Add(ptr, b.addr+off)
	mem := unsafe.Slice((*byte)(ptr), len(buf))

	b.copy(mem, buf, 8, off, true)
}

func (b *block) slice() (buf []byte) {
	var ptr unsafe.Pointer

	ptr = unsafe.Add(ptr, b.addr)
	buf = unsafe.Slice((*byte)(ptr), b.size)

	return
}
