// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package bits

// Get16 returns whether a specific bit position is set at the pointed value.
func Get16(addr *uint16, pos int) bool {
	return (int(*addr)>>pos)&1 == 1
}

// Set16 modifies the pointed value by setting an individual bit at the
// position argument.
func Set16(addr *uint16, pos int) {
	*addr |= (1 << pos)
}

// SetTo16 modifies the pointed value by setting an individual bit at the
// position argument.
func SetTo16(addr *uint16, pos int, val bool) {
	if val {
		Set16(addr, pos)
	} else {
		Clear16(addr, pos)
	}
}

// Clear16 modifies the pointed value by clearing an individual bit at the
// position argument.
func Clear16(addr *uint16, pos int) {
	*addr &= ^(1 << pos)
}

// GetN16 returns the pointed value at a specific bit position and with a
// bitmask applied.
func GetN16(addr *uint16, pos int, mask int) uint16 {
	return uint16((int(*addr) >> pos) & mask)
}

// SetN16 modifies the pointed value by setting a value at a specific bit
// position and with a bitmask applied.
func SetN16(addr *uint16, pos int, mask int, val uint16) {
	*addr = (*addr & (^(uint16(mask) << pos))) | (val << pos)
}
