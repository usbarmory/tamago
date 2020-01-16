// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package mem provides primitives for data structure memory allocation and
// alignment.
package mem

import (
	"unsafe"
)

// AlignmentBuffer provides a byte array buffer for data structure alignment
// purposes.
type AlignmentBuffer struct {
	Offset int
	Buf  []byte
}

// NewAlignmentBuffer initializes a buffer and offset to achieve the requested
// alignment, such as for allocating aligned structures by casting them over
// the buffer offset.
func NewAlignmentBuffer(size uintptr, align int) (ab *AlignmentBuffer) {
	ab = &AlignmentBuffer{}
	buf := make([]byte, int(size)+align)

	ab.Buf = buf
	addr := int(uintptr(unsafe.Pointer(&buf[0])))

	if align <= 0 {
		return
	}

	if check(addr, align) {
		return
	}

	if r := addr & (align - 1); r != 0 {
		ab.Offset = (align - r)
		addr += ab.Offset
	}

	if !check(addr, align) {
		panic("alignment error\n")
	}

	return
}

// Addr returns the memory address corresponding to the aligned buffer offset.
func (ab *AlignmentBuffer) Addr() uint32 {
	return uint32(uintptr(unsafe.Pointer(&ab.Buf[ab.Offset])))
}

// Ptr returns a pointer to the memory address corresponding to the aligned
// buffer offset.
func (ab *AlignmentBuffer) Ptr() unsafe.Pointer {
	return unsafe.Pointer(&ab.Buf[ab.Offset])
}

// Data returns the aligned data stored in the buffer.
func (ab *AlignmentBuffer) Data() []byte {
	return ab.Buf[ab.Offset:]
}

// Fill copies a byte array to an aligned buffer.
func Copy(ab *AlignmentBuffer, data []byte) {
	copy(ab.Buf[ab.Offset:], data)
}

func check(addr int, align int) bool {
	return addr&(align-1) == 0
}
