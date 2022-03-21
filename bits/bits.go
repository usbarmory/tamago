// https://github.com/usbarmory/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package bits provides primitives for bitwise operations on uint32 values.
package bits

// Get returns the pointed value at a specific bit position and with a bitmask
// applied.
func Get(addr *uint32, pos int, mask int) uint32 {
	return uint32((int(*addr) >> pos) & mask)
}

// Set modifies the pointed value by setting an individual bit at the position
// argument.
func Set(addr *uint32, pos int) {
	*addr |= (1 << pos)
}

// Clear modifies the pointed value by clearing an individual bit at the
// position argument.
func Clear(addr *uint32, pos int) {
	*addr &= ^(1 << pos)
}

// SetN modifies the pointed value by setting a value at a specific bit
// position and with a bitmask applied.
func SetN(addr *uint32, pos int, mask int, val uint32) {
	*addr = (*addr & (^(uint32(mask) << pos))) | (val << pos)
}
