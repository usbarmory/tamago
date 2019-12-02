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

func Wait(reg *uint32, pos int, mask int, val uint32) {
	for Get(reg, pos, mask) != val {
	}
}

func WaitFor(timeout time.Duration, reg *uint32, pos int, mask int, val uint32) bool {
	start := time.Now()

	for Get(reg, pos, mask) != val {
		if time.Since(start) >= timeout {
			return false
		}
	}

	return true
}
