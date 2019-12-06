// https://github.com/inversepath/tamago
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
	"time"

	"github.com/inversepath/tamago/imx6/internal/cache"
)

func Get(reg *uint32, pos int, mask int) uint32 {
	// TODO: disable cache for peripheral space
	cache.FlushData()
	return uint32((int(*reg) >> pos) & mask)
}

func Set(reg *uint32, pos int) {
	*reg |= (1 << pos)
}

func Clear(reg *uint32, pos int) {
	*reg &= ^(1 << pos)
}

func SetN(reg *uint32, pos int, mask int, val uint32) {
	*reg = (*reg & (^(uint32(mask) << pos))) | (val << pos)
}

func ClearN(reg *uint32, pos int, mask int) {
	*reg &= ^(uint32(mask) << pos)
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
