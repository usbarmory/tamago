// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package iram provides primitives for data structure memory allocation on the
// i.MX6 internal OCRAM (iRAM).
//
// It is provided for bare metal device driver operation, to avoid passing Go
// pointers for DMA purposes.
//
// The iRAM (128KB) is used with an extremely simple allocation scheme, blocks
// of memory can be allocated for a given size, using a string tag identifier.
// These blocks can be re-used with the same tag and for less or equal
// allocations, they cannot be reclaimed or deallocated, but only zeroed.
//
// For the current required device drivers this is perfectly adequate, moving
// to a more complex allocation, or replacing iRAM with RAM based scratch space
// if size becomes a constraint for DMA transfers, will be performed as needed.
package iram

import (
	"unsafe"
)

const iramStart uint32 = 0x00900000
const iramSize uint32 = 0x20000

type block struct {
	start uint32
	end   uint32
	size  int
}

var blocks map[string]*block
var lastFree uint32

func init() {
	blocks = make(map[string]*block)
	lastFree = iramStart
}

func (b *block) read() []byte {
	data := make([]byte, b.size)

	for i := 0; i < b.size; i++ {
		data[i] = *(*byte)(unsafe.Pointer(uintptr(b.start + uint32(i))))
	}

	return data
}

func (b *block) write(data []byte) {
	b.size = len(data)

	for i := range data {
		*(*byte)(unsafe.Pointer(uintptr(b.start + uint32(i)))) = data[i]
	}
}

func (b *block) free() {
	for i := 0; i < b.size; i++ {
		*(*byte)(unsafe.Pointer(uintptr(b.start + uint32(i)))) = 0x00
	}
}

// Read returns a byte array from previously allocated internal memory.
func Read(tag string) (data []byte) {
	if b, ok := blocks[tag]; ok {
		data = b.read()
	}

	return
}

// Write copies a byte array in internal memory and return its allocation
// pointer, with optional alignment, a string tag allows to re-use the
// previously allocated memory slot as long as data size is less or equal than
// what previously allocated.
func Write(tag string, data []byte, align int) uint32 {
	var b *block
	var exists bool

	size := len(data)

	if b, exists = blocks[tag]; exists {
		if size > int(b.end - b.start) {
			panic("attempt to re-use slot with larger size (" + tag + ")")
		}

		if align != 0 && !check(int(b.start), align) {
			panic("attempt to re-use slot with different alignment")
		}

		b.size = size
	} else {
		start := int(lastFree)

		if align > 0 && !check(start, align) {
			if r := start & (align - 1); r != 0 {
				start += (align - r)
			}
		}

		if uint32(start + size) >= (iramStart + iramSize) {
			panic("out of iram memory")
		}

		b = &block{
			start: uint32(start),
			end: uint32(start + size),
			size: size,
		}

		blocks[tag] = b
		lastFree = b.start + uint32(size)
	}

	b.write(data)

	return b.start
}

// Free clears a previously allocated internal memory block by zeroing out its
// contents, memory is not reclaimed but can be reused for an identical tag of
// less or equal size.
func Free(tag string) {
	if b, ok := blocks[tag]; ok {
		b.free()
	}
}

func check(addr int, align int) bool {
	return addr&(align-1) == 0
}
