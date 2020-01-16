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
// to a more complex allocation will be performed only if required.
package iram

import (
	"unsafe"
)

const iramStart uint32 = 0x00900000
const iramSize uint32 = 0x20000

type block struct {
	addr uint32
	size int
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
		data[i] = *(*byte)(unsafe.Pointer(uintptr(b.addr + uint32(i))))
	}

	return data
}

func (b *block) write(data []byte) {
	b.size = len(data)

	for i := range data {
		*(*byte)(unsafe.Pointer(uintptr(b.addr + uint32(i)))) = data[i]
	}
}

func (b *block) free() {
	for i := 0; i < b.size; i++ {
		*(*byte)(unsafe.Pointer(uintptr(b.addr + uint32(i)))) = 0x00
	}
}

// Read returns a byte array from previously allocated internal memory.
func Read(tag string) (data []byte) {
	if b, ok := blocks[tag]; ok {
		data = b.read()
	}

	return
}

// Copy copies a byte array in internal memory and return its allocation
// pointer, a string tag allows to re-use the previously allocated memory slot.
func Write(tag string, data []byte) uint32 {
	var b *block
	var exists bool

	size := len(data)

	if b, exists = blocks[tag]; exists {
		if size > b.size {
			panic("attempt to re-use slot but with larger size (" + tag + ")")
		}

		b.size = size
	} else {
		if lastFree+uint32(size) >= (iramStart + iramSize) {
			panic("out of iram memory")
		}

		b = &block{
			addr: lastFree,
			size: size,
		}

		blocks[tag] = b
		lastFree += uint32(size)
	}

	b.write(data)

	return b.addr
}

// Free clears a previously allocated internal memory block by zeroing out its
// contents, memory is not reclaimed but can be reused for an identical tag of
// less or equal size.
func Free(tag string) {
	if b, ok := blocks[tag]; ok {
		b.free()
	}
}
