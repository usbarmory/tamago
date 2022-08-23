// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package bits

// Get64 returns the pointed value at a specific bit position and with a
// bitmask applied.
func Get64(addr *uint64, pos int, mask int) uint64 {
	return uint64((int(*addr) >> pos) & mask)
}

// Set64 modifies the pointed value by setting an individual bit at the
// position argument.
func Set64(addr *uint64, pos int) {
	*addr |= (1 << pos)
}

// Clear64 modifies the pointed value by clearing an individual bit at the
// position argument.
func Clear64(addr *uint64, pos int) {
	*addr &= ^(1 << pos)
}

// SetTo64 modifies the pointed value by setting an individual bit at the
// position argument.
func SetTo64(addr *uint64, pos int, val bool) {
	if val {
		Set64(addr, pos)
	} else {
		Clear64(addr, pos)
	}
}

// SetN64 modifies the pointed value by setting a value at a specific bit
// position and with a bitmask applied.
func SetN64(addr *uint64, pos int, mask int, val uint64) {
	*addr = (*addr & (^(uint64(mask) << pos))) | (val << pos)
}
