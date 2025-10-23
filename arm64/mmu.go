// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm64

// MMU access permissions
// (Table G5-9, ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile).
const (
	// PL1: no access   PL0: no access
	TTE_AP_000 uint32 = 0b00
	// PL1: read/write  PL0: no access
	TTE_AP_001 uint32 = 0b01
	// PL1: read/write  PL0: read only
	TTE_AP_010 uint32 = 0b10
	// PL1: read/write  PL0: read/write
	TTE_AP_011 uint32 = 0b11
)

// InitMMU initializes the first-level translation tables for all available
// memory with a flat mapping and privileged attribute flags.
//
// The first 4096 bytes (0x00000000 - 0x00001000) are flagged as invalid to
// trap null pointers.
//
// All available memory is marked as non-executable except for the range
// returned by runtime.TextRegion().
func (cpu *CPU) InitMMU() {
	return // TODO
}
