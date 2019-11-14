// https://github.com/inversepath/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package mem

import (
	"fmt"
	"unsafe"
)

func AlignedBuffer(size uintptr, align uintptr) (buf []byte, addr uintptr, err error) {
	buf = make([]byte, size+align)
	addr = uintptr(unsafe.Pointer(&buf[0]))

	if r := addr & (align - 1); r != 0 {
		addr += (align - r)
	}

	if !isAligned(addr, align) {
		err = fmt.Errorf("buffer alignment failed, addr:%x, align:%x", addr, align)
	}

	return
}

func isAligned(addr uintptr, align uintptr) bool {
	if addr&(align-1) == 0 {
		return true
	}

	return false
}
