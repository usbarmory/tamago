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

// As sync/atomic does not provide 8-bit support, note that these functions do
// not necessarily enforce memory ordering.

func Get8(addr uint32, pos int) bool {
	reg := (*uint8)(unsafe.Pointer(uintptr(addr)))
	return (*reg>>pos)&1 == 1
}

func Set8(addr uint32, pos int) {
	reg := (*uint8)(unsafe.Pointer(uintptr(addr)))
	*reg |= (1 << pos)
}

func SetTo8(addr uint32, pos int, val bool) {
	if val {
		Set8(addr, pos)
	} else {
		Clear8(addr, pos)
	}
}

func Clear8(addr uint32, pos int) {
	reg := (*uint8)(unsafe.Pointer(uintptr(addr)))
	*reg &= ^(uint8(1) << pos)
}

func GetN8(addr uint32, pos int, mask int) uint8 {
	reg := (*uint8)(unsafe.Pointer(uintptr(addr)))
	return (*reg >> pos) & uint8(mask)
}

func SetN8(addr uint32, pos int, mask int, val uint8) {
	reg := (*uint8)(unsafe.Pointer(uintptr(addr)))
	*reg = (*reg & (^(uint8(mask) << pos))) | (val << pos)
}

func ClearN8(addr uint32, pos int, mask int) {
	reg := (*uint8)(unsafe.Pointer(uintptr(addr)))
	*reg &= ^(uint8(mask) << pos)
}

func Read8(addr uint32) uint8 {
	reg := (*uint8)(unsafe.Pointer(uintptr(addr)))
	return *reg
}

func Write8(addr uint32, val uint8) {
	reg := (*uint8)(unsafe.Pointer(uintptr(addr)))
	*reg = val
}

// Wait8 waits for a specific register bit to match a value. This function
// cannot be used before runtime initialization with `GOOS=tamago`.
func Wait8(addr uint32, pos int, mask int, val uint8) {
	for GetN8(addr, pos, mask) != val {
		runtime.Gosched()
	}
}

// WaitFor8 waits, until a timeout expires, for a specific register bit to match
// a value. The return boolean indicates whether the wait condition was checked
// (true) or if it timed out (false). This function cannot be used before
// runtime initialization with `GOOS=tamago`.
func WaitFor8(timeout time.Duration, addr uint32, pos int, mask int, val uint8) bool {
	start := time.Now()

	for GetN8(addr, pos, mask) != val {
		runtime.Gosched()

		if time.Since(start) >= timeout {
			return false
		}
	}

	return true
}
