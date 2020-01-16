// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package bits provides primitives for bitwise operations on uint32 values.
package bits

import (
	"runtime"

	"github.com/f-secure-foundry/tamago/imx6/internal/cache"
)

func Get(addr *uint32, pos int, mask int) uint32 {
	return uint32((int(*addr) >> pos) & mask)
}

func Set(addr *uint32, pos int) {
	*addr |= (1 << pos)
}

func Clear(addr *uint32, pos int) {
	*addr &= ^(1 << pos)
}

func SetN(addr *uint32, pos int, mask int, val uint32) {
	*addr = (*addr & (^(uint32(mask) << pos))) | (val << pos)
}

// Wait waits for a specific register bit to match a value. This function
// cannot be used before runtime initialization with `GOOS=tamago`.
func Wait(addr *uint32, pos int, mask int, val uint32) {
	cache.FlushData()

	for Get(addr, pos, mask) != val {
		// tamago is single-threaded so we must force giving
		// other goroutines a chance
		runtime.Gosched()
		cache.FlushData()
	}
}
