// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build tamago,arm

// Package reg provides primitives for retrieving and modifying hardware
// registers.
package reg

import (
	"runtime"
	"sync"
	"time"
	"unsafe"

	"github.com/f-secure-foundry/tamago/arm"
)

// TODO: disable cache for peripheral space instead of cache.FlushData() every
// time.

var mutex sync.Mutex

func Get(addr uint32, pos int, mask int) (val uint32) {
	reg := (*uint32)(unsafe.Pointer(uintptr(addr)))

	mutex.Lock()

	arm.CacheFlushData()
	val = uint32((int(*reg) >> pos) & mask)

	mutex.Unlock()

	return
}

func Set(addr uint32, pos int) {
	reg := (*uint32)(unsafe.Pointer(uintptr(addr)))

	mutex.Lock()

	arm.CacheFlushData()
	*reg |= (1 << pos)

	mutex.Unlock()
}

func Clear(addr uint32, pos int) {
	reg := (*uint32)(unsafe.Pointer(uintptr(addr)))

	mutex.Lock()

	arm.CacheFlushData()
	*reg &= ^(1 << pos)

	mutex.Unlock()
}

func SetN(addr uint32, pos int, mask int, val uint32) {
	reg := (*uint32)(unsafe.Pointer(uintptr(addr)))

	mutex.Lock()

	arm.CacheFlushData()
	*reg = (*reg & (^(uint32(mask) << pos))) | (val << pos)

	mutex.Unlock()
}

func ClearN(addr uint32, pos int, mask int) {
	reg := (*uint32)(unsafe.Pointer(uintptr(addr)))

	mutex.Lock()

	arm.CacheFlushData()
	*reg &= ^(uint32(mask) << pos)

	mutex.Unlock()
}

func Read(addr uint32) (val uint32) {
	reg := (*uint32)(unsafe.Pointer(uintptr(addr)))

	mutex.Lock()

	arm.CacheFlushData()
	val = *reg

	mutex.Unlock()

	return
}

func Write(addr uint32, val uint32) {
	reg := (*uint32)(unsafe.Pointer(uintptr(addr)))

	mutex.Lock()

	arm.CacheFlushData()
	*reg = val

	mutex.Unlock()
}

func WriteBack(addr uint32) {
	reg := (*uint32)(unsafe.Pointer(uintptr(addr)))

	mutex.Lock()

	arm.CacheFlushData()
	*reg |= *reg

	mutex.Unlock()
}

func Or(addr uint32, val uint32) {
	reg := (*uint32)(unsafe.Pointer(uintptr(addr)))

	mutex.Lock()

	arm.CacheFlushData()
	*reg |= val

	mutex.Unlock()
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
// runtime initialization with `GOOS=tamago`.
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
