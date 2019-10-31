// https://github.com/inversepath/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package imx6

import (
	"fmt"
	"unsafe"
)

func get(reg *uint32, pos int, mask int) uint32 {
	return uint32((int(*reg) >> pos) & mask)
}

func set(reg *uint32, pos int) {
	*reg |= (1 << pos)
}

func clear(reg *uint32, pos int) {
	*reg &= ^(1 << pos)
}

func setN(reg *uint32, pos int, mask int, val uint32) {
	*reg = (*reg & (^(uint32(mask) << pos))) | (val << pos)
}

func clearN(reg *uint32, pos int, mask int) {
	*reg &= ^(uint32(mask) << pos)
}

func wait(reg *uint32, pos int, mask int, val uint32) {
	for get(reg, pos, mask) != val {
	}
}

func alignedBuffer(size uintptr, align uintptr) (buf []byte, addr uintptr, err error) {
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
