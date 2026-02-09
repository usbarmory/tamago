// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package bits provides primitives for bitwise operations on 32/64 bit
// registers.
package bits

// Get returns whether a specific bit position is set at the pointed value.
func Get(addr *uint32, pos int) bool {
	return (int(*addr)>>pos)&1 == 1
}

// Set modifies the pointed value by setting an individual bit at the position
// argument.
func Set(addr *uint32, pos int) {
	*addr |= (1 << pos)
}

// SetTo modifies the pointed value by setting an individual bit at the
// position argument.
func SetTo(addr *uint32, pos int, val bool) {
	if val {
		Set(addr, pos)
	} else {
		Clear(addr, pos)
	}
}

// Clear modifies the pointed value by clearing an individual bit at the
// position argument.
func Clear(addr *uint32, pos int) {
	*addr &= ^(1 << pos)
}

// GetN returns the pointed value at a specific bit position and with a bitmask
// applied.
func GetN(addr *uint32, pos int, mask int) uint32 {
	return uint32((int(*addr) >> pos) & mask)
}

// SetN modifies the pointed value by setting a value at a specific bit
// position and with a bitmask applied.
func SetN(addr *uint32, pos int, mask int, val uint32) {
	*addr = (*addr & (^(uint32(mask) << pos))) | (val << pos)
}
