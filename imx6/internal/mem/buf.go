// https://github.com/inversepath/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package mem

import (
	"unsafe"
)

type AlignmentBuffer struct {
	Addr uintptr
	Offset uintptr
	Buf  []byte
}

// Return a buffer and offset that allow to achieve the desired alignment, such
// as for allocating aligned structures by casting them over the buffer offset.
// It is important to keep the []byte buffer around until the cast object is
// required, to avoid GC swiping it away (as we go through uintptr).
func (ab *AlignmentBuffer) Init(size uintptr, align uintptr) {
	buf := make([]byte, size+align)

	ab.Buf = buf
	ab.Addr = uintptr(unsafe.Pointer(&buf[0]))

	if ab.check(align) {
		return
	}

	if r := ab.Addr & (align - 1); r != 0 {
		ab.Offset = (align - r)
		ab.Addr += ab.Offset
	}

	if !ab.check(align) {
		panic("alignment error\n")
	}
}

func (ab *AlignmentBuffer) check(align uintptr) bool {
	return ab.Addr&(align-1) == 0
}

// Copy copies a byte array to an aligned buffer.
func Copy(ab AlignmentBuffer, data []byte) {
	copy(ab.Buf[ab.Offset:], data)
}
