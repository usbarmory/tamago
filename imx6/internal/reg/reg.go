// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package reg provides primitves for retrieving and modifying hardware
// registers.
package reg

import (
	"runtime"
	"sync"
	"time"

	"github.com/f-secure-foundry/tamago/imx6/internal/cache"
)

// TODO: disable cache for peripheral space instead of cache.FlushData() every
// time

var mutex sync.Mutex

func Get(reg *uint32, pos int, mask int) (val uint32) {
	mutex.Lock()

	cache.FlushData()
	val = uint32((int(*reg) >> pos) & mask)

	mutex.Unlock()

	return
}

func Set(reg *uint32, pos int) {
	mutex.Lock()

	cache.FlushData()
	*reg |= (1 << pos)

	mutex.Unlock()
}

func Write(reg *uint32, val uint32) {
	mutex.Lock()

	cache.FlushData()
	*reg = val

	mutex.Unlock()
}

func Clear(reg *uint32, pos int) {
	mutex.Lock()

	cache.FlushData()
	*reg &= ^(1 << pos)

	mutex.Unlock()
}

func SetN(reg *uint32, pos int, mask int, val uint32) {
	mutex.Lock()

	cache.FlushData()
	*reg = (*reg & (^(uint32(mask) << pos))) | (val << pos)

	mutex.Unlock()
}

func ClearN(reg *uint32, pos int, mask int) {
	mutex.Lock()

	cache.FlushData()
	*reg &= ^(uint32(mask) << pos)

	mutex.Unlock()
}

// Wait waits for a specific register bit to match a value. This function
// cannot be used before runtime initialization with `GOOS=tamago`.
func Wait(reg *uint32, pos int, mask int, val uint32) {
	for Get(reg, pos, mask) != val {
		// tamago is single-threaded so we must force giving
		// other goroutines a chance
		runtime.Gosched()
	}
}

// WaitFor waits, until a timeout expires, for a specific register bit to match
// a value. The return boolean indicates whether the wait condition was checked
// (true) or if it timed out (false). This function cannot be used before
// runtime initialization with `GOOS=tamago`.
func WaitFor(timeout time.Duration, reg *uint32, pos int, mask int, val uint32) bool {
	start := time.Now()

	for Get(reg, pos, mask) != val {
		// tamago is single-threaded so we must force giving
		// other goroutines a chance
		runtime.Gosched()

		if time.Since(start) >= timeout {
			return false
		}
	}

	return true
}
