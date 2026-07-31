// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package reg

import (
	"runtime"
	"time"
	"unsafe"
)

// As sync/atomic does not provide 16-bit support, note that these functions do
// not necessarily enforce memory ordering.

func Get16(addr uint32, pos int) bool {
	reg := (*uint16)(unsafe.Pointer(uintptr(addr)))
	r := *reg

	return (r>>pos)&1 == 1
}

func Set16(addr uint32, pos int) {
	reg := (*uint16)(unsafe.Pointer(uintptr(addr)))
	*reg |= (1 << pos)
}

func SetTo16(addr uint32, pos int, val bool) {
	if val {
		Set16(addr, pos)
	} else {
		Clear16(addr, pos)
	}
}

func Clear16(addr uint32, pos int) {
	reg := (*uint16)(unsafe.Pointer(uintptr(addr)))
	*reg &= ^(1 << pos)
}

func GetN16(addr uint32, pos int, mask int) uint16 {
	reg := (*uint16)(unsafe.Pointer(uintptr(addr)))
	return (*reg >> pos) & uint16(mask)
}

func SetN16(addr uint32, pos int, mask int, val uint16) {
	reg := (*uint16)(unsafe.Pointer(uintptr(addr)))
	*reg = (*reg & (^(uint16(mask) << pos))) | ((val & uint16(mask)) << pos)
}

func ClearN16(addr uint32, pos int, mask int) {
	reg := (*uint16)(unsafe.Pointer(uintptr(addr)))
	*reg &= ^(uint16(mask) << pos)
}

func Read16(addr uint32) uint16 {
	reg := (*uint16)(unsafe.Pointer(uintptr(addr)))
	return *reg
}

func Write16(addr uint32, val uint16) {
	reg := (*uint16)(unsafe.Pointer(uintptr(addr)))
	*reg = val
}

func Or16(addr uint32, val uint16) {
	reg := (*uint16)(unsafe.Pointer(uintptr(addr)))
	*reg |= val
}

func Wait16(addr uint32, pos int, mask int, val uint16) {
	for GetN16(addr, pos, mask) != val {
		runtime.Gosched()
	}
}

func WaitFor16(timeout time.Duration, addr uint32, pos int, mask int, val uint16) bool {
	start := time.Now()

	for GetN16(addr, pos, mask) != val {
		runtime.Gosched()

		if time.Since(start) >= timeout {
			return false
		}
	}

	return true
}
