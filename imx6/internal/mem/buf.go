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

// Return a buffer and offset that allow to achieve the desired alignment, such
// as for allocating aligned structures by casting them over the buffer offset.
// It is important to keep the []byte buffer around until the cast object is
// required, to avoid GC swiping it away (as we go through uintptr).
func AlignedBuffer(size uintptr, align uintptr) (*[]byte, uintptr) {
	buf := make([]byte, size+align)
	addr := uintptr(unsafe.Pointer(&buf[0]))

	if IsAligned(addr, align) {
		return &buf, addr
	}

	if r := addr & (align - 1); r != 0 {
		addr += (align - r)
	}

	if !IsAligned(addr, align) {
		panic("alignment error\n")
	}

	return &buf, addr
}

func IsAligned(addr uintptr, align uintptr) bool {
	if addr&(align-1) == 0 {
		return true
	}

	return false
}
