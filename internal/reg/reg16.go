// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
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

func Get16(addr uint32, pos int, mask int) uint16 {
	reg := (*uint16)(unsafe.Pointer(uintptr(addr)))
	return (*reg >> pos) & uint16(mask)
}

func Set16(addr uint32, pos int) {
	reg := (*uint16)(unsafe.Pointer(uintptr(addr)))
	*reg |= (1 << pos)
}

func Clear16(addr uint32, pos int) {
	reg := (*uint16)(unsafe.Pointer(uintptr(addr)))
	*reg &= ^(1 << pos)
}

func SetTo16(addr uint32, pos int, val bool) {
	if val {
		Set16(addr, pos)
	} else {
		Clear16(addr, pos)
	}
}

func SetN16(addr uint32, pos int, mask int, val uint16) {
	reg := (*uint16)(unsafe.Pointer(uintptr(addr)))
	*reg = (*reg & (^(uint16(mask) << pos))) | (val << pos)
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

func WriteBack16(addr uint32) {
	reg := (*uint16)(unsafe.Pointer(uintptr(addr)))
	*reg |= *reg
}

func Or16(addr uint32, val uint16) {
	reg := (*uint16)(unsafe.Pointer(uintptr(addr)))
	*reg |= val
}

// Wait16 waits for a specific register bit to match a value. This function
// cannot be used before runtime initialization with `GOOS=tamago`.
func Wait16(addr uint32, pos int, mask int, val uint16) {
	for Get16(addr, pos, mask) != val {
		// tamago is single-threaded, give other goroutines a chance
		runtime.Gosched()
	}
}

// WaitFor16 waits, until a timeout expires, for a specific register bit to match
// a value. The return boolean indicates whether the wait condition was checked
// (true) or if it timed out (false). This function cannot be used before
// runtime initialization with `GOOS=tamago`.
func WaitFor16(timeout time.Duration, addr uint32, pos int, mask int, val uint16) bool {
	start := time.Now()

	for Get16(addr, pos, mask) != val {
		// tamago is single-threaded, give other goroutines a chance
		runtime.Gosched()

		if time.Since(start) >= timeout {
			return false
		}
	}

	return true
}
