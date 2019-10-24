// https://github.com/inversepath/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package imx6

import (
	"errors"
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
	off := uintptr(0)
	buf = make([]byte, size+align)
	addr = uintptr(unsafe.Pointer(&buf[off]))

	if r := addr & (4 - 1); r != 0 {
		off += r
		addr += off
	}

	if !isAligned(addr, align) {
		err = errors.New("buffer alignment failed")
	}

	return
}

func isAligned(p uintptr, n uintptr) bool {
	if p&(n-1) == 0 {
		return true
	}

	return false
}
