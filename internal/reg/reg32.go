// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package reg provides primitives for retrieving and modifying hardware
// registers.
//
// This package is only meant to be used with `GOOS=tamago` as supported by the
// TamaGo framework for bare metal Go on ARM/RISC-V SoCs, see
// https://github.com/usbarmory/tamago.
package reg

import (
	"runtime"
	"sync/atomic"
	"time"
	"unsafe"
)

func IsSet(addr uint32, pos int) bool {
	reg := (*uint32)(unsafe.Pointer(uintptr(addr)))
	r := atomic.LoadUint32(reg)

	return (int(r) >> pos) & 1 == 1
}

func Get(addr uint32, pos int, mask int) uint32 {
	reg := (*uint32)(unsafe.Pointer(uintptr(addr)))
	r := atomic.LoadUint32(reg)

	return uint32((int(r) >> pos) & mask)
}

func Set(addr uint32, pos int) {
	reg := (*uint32)(unsafe.Pointer(uintptr(addr)))

	r := atomic.LoadUint32(reg)
	r |= (1 << pos)

	atomic.StoreUint32(reg, r)
}

func Clear(addr uint32, pos int) {
	reg := (*uint32)(unsafe.Pointer(uintptr(addr)))

	r := atomic.LoadUint32(reg)
	r &= ^(1 << pos)

	atomic.StoreUint32(reg, r)
}

func SetTo(addr uint32, pos int, val bool) {
	if val {
		Set(addr, pos)
	} else {
		Clear(addr, pos)
	}
}

func SetN(addr uint32, pos int, mask int, val uint32) {
	reg := (*uint32)(unsafe.Pointer(uintptr(addr)))

	r := atomic.LoadUint32(reg)
	r = (r & (^(uint32(mask) << pos))) | (val << pos)

	atomic.StoreUint32(reg, r)
}

func ClearN(addr uint32, pos int, mask int) {
	reg := (*uint32)(unsafe.Pointer(uintptr(addr)))

	r := atomic.LoadUint32(reg)
	r &= ^(uint32(mask) << pos)

	atomic.StoreUint32(reg, r)
}

// defined in reg32_*.s
func Move(dst uint32, src uint32)

func Read(addr uint32) uint32 {
	reg := (*uint32)(unsafe.Pointer(uintptr(addr)))
	return atomic.LoadUint32(reg)
}

func Write(addr uint32, val uint32) {
	reg := (*uint32)(unsafe.Pointer(uintptr(addr)))
	atomic.StoreUint32(reg, val)
}

func WriteBack(addr uint32) {
	reg := (*uint32)(unsafe.Pointer(uintptr(addr)))

	r := atomic.LoadUint32(reg)
	r |= r

	atomic.StoreUint32(reg, r)
}

func Or(addr uint32, val uint32) {
	reg := (*uint32)(unsafe.Pointer(uintptr(addr)))

	r := atomic.LoadUint32(reg)
	r |= val

	atomic.StoreUint32(reg, r)
}

// Wait waits for a specific register bit to match a value. This function
// cannot be used before runtime initialization with `GOOS=tamago`.
func Wait(addr uint32, pos int, mask int, val uint32) {
	for Get(addr, pos, mask) != val {
		// tamago is single-threaded, give other goroutines a chance
		runtime.Gosched()
	}
}

// WaitFor waits, until a timeout expires, for a specific register bit to match
// a value. The return boolean indicates whether the wait condition was checked
// (true) or if it timed out (false). This function cannot be used before
// runtime initialization.
func WaitFor(timeout time.Duration, addr uint32, pos int, mask int, val uint32) bool {
	start := time.Now()

	for Get(addr, pos, mask) != val {
		// tamago is single-threaded, give other goroutines a chance
		runtime.Gosched()

		if time.Since(start) >= timeout {
			return false
		}
	}

	return true
}

// WaitSignal waits, until a channel is closed, for a specific register bit to
// match a value. The return boolean indicates whether the wait condition was
// checked (true) or cancelled (false). This function cannot be used before
// runtime initialization.
func WaitSignal(exit chan struct{}, addr uint32, pos int, mask int, val uint32) bool {
	for Get(addr, pos, mask) != val {
		// tamago is single-threaded, give other goroutines a chance
		runtime.Gosched()

		select {
		case <-exit:
			return false
		default:
		}
	}

	return true
}
