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

type block struct {
	// pointer address
	addr uint
	// buffer size
	size uint
	// distinguish regular (`Alloc`/`Free`) and reserved
	// (`Reserve`/`Release`) blocks.
	res bool
}

func (b *block) read(off uint, buf []byte) {
	var ptr unsafe.Pointer

	ptr = unsafe.Add(ptr, b.addr+off)
	mem := unsafe.Slice((*byte)(ptr), len(buf))

	copy(buf, mem)
}

func (b *block) write(off uint, buf []byte) {
	var ptr unsafe.Pointer

	ptr = unsafe.Add(ptr, b.addr+off)
	mem := unsafe.Slice((*byte)(ptr), len(buf))

	copy(mem, buf)
}

func (b *block) slice() (buf []byte) {
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	hdr.Data = uintptr(unsafe.Pointer(uintptr(b.addr)))
	hdr.Len = int(b.size)
	hdr.Cap = hdr.Len

	return
}
