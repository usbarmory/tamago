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
// not enforce memory ordering at the CPU level, as no barriers are issued.
//
// All accesses are performed through Read16 and Write16, which must not be
// inlined to ensure that each access is issued exactly once, with 16-bit width
// and in program order, as the compiler might otherwise combine or eliminate
// plain memory accesses (e.g. two consecutive reads of the same register).

func Get16(addr uint32, pos int) bool {
	return (Read16(addr)>>pos)&1 == 1
}

func Set16(addr uint32, pos int) {
	Write16(addr, Read16(addr)|(1<<pos))
}

func SetTo16(addr uint32, pos int, val bool) {
	if val {
		Set16(addr, pos)
	} else {
		Clear16(addr, pos)
	}
}

func Clear16(addr uint32, pos int) {
	Write16(addr, Read16(addr)&^(1<<pos))
}

func GetN16(addr uint32, pos int, mask int) uint16 {
	return (Read16(addr) >> pos) & uint16(mask)
}

func SetN16(addr uint32, pos int, mask int, val uint16) {
	r := Read16(addr)
	Write16(addr, (r&(^(uint16(mask) << pos)))|((val&uint16(mask))<<pos))
}

func ClearN16(addr uint32, pos int, mask int) {
	Write16(addr, Read16(addr)&^(uint16(mask)<<pos))
}

//go:noinline
func Read16(addr uint32) uint16 {
	reg := (*uint16)(unsafe.Pointer(uintptr(addr)))
	return *reg
}

//go:noinline
func Write16(addr uint32, val uint16) {
	reg := (*uint16)(unsafe.Pointer(uintptr(addr)))
	*reg = val
}

func Or16(addr uint32, val uint16) {
	Write16(addr, Read16(addr)|val)
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
