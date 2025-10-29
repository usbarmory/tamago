// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm64

import (
	"runtime"

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
)

const (
	l1pageTableOffset = 0x4000
	l1pageTableSize   = 4096

	l2pageTableOffset = 0xc000
	l2pageTableSize   = 512
)

// Memory region attributes (Table 4-113 ARM Architecture Reference Manual
// ARMv8, for ARMv8-A architecture profile).
const (
	TTE_PAGE_TABLE uint64 = (1 << 0)

	MemoryRegion uint64 = 0b11111111
	DeviceRegion uint64 = 0b00000000
)

// MMU access permissions (Table G5-9, ARM Architecture Reference Manual ARMv8,
// for ARMv8-A architecture profile).
const (
	// PL1: no access   PL0: no access
	TTE_AP_000 uint64 = 0b00
	// PL1: read/write  PL0: no access
	TTE_AP_001 uint64 = 0b01
	// PL1: read/write  PL0: read only
	TTE_AP_010 uint64 = 0b10
	// PL1: read/write  PL0: read/write
	TTE_AP_011 uint64 = 0b11
)

// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile
// D12.2.105 TCR_EL3, Translation Control Register (EL3).
const (
	TCR_PS    = 16
	TCR_TG0   = 14
	TCR_SH0   = 12
	TCR_ORGN0 = 8
	TCR_IRGN0 = 6
	TCR_T0SZ  = 5
)

// defined in mmu.s
func write_mair_el3(val uint64)
func write_tcr_el3(val uint32)
func set_ttbr0_el3(addr uint64)

// D5.3.1 Translation table level 0, level 1, and level 2 descriptor formats
// (ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile).
func (cpu *CPU) initL1Table(entry int, ttbr uint64, section uint64) {
}

// D5.3.1 Translation table level 0, level 1, and level 2 descriptor formats
// (ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile).
func (cpu *CPU) initL2Table(entry int, base uint64, section uint64) {
}

// InitMMU initializes the first-level translation tables for all available
// memory with a flat mapping and privileged attribute flags.
//
// The first 4096 bytes (0x00000000 - 0x00001000) are flagged as invalid to
// trap null pointers.
//
// All available memory is marked as non-executable except for the range
// returned by runtime.TextRegion().
func (cpu *CPU) InitMMU() {
	ramStart, _ := runtime.MemRegion()
	l1pageTableStart := ramStart + l1pageTableOffset
	l2pageTableStart := ramStart + l2pageTableOffset

	// Map the first L1 entry to an L2 table to trap null pointers within
	// the smallest possible section (4KB starting from 0x00000000).
	firstSection := l2pageTableStart | TTE_PAGE_TABLE
	reg.Write64(l1pageTableStart, firstSection)

	// set first L2 entry as invalid
	reg.Write64(l2pageTableStart, 0)

	// set remaining entries with flat mapping
	cpu.initL1Table(1, l1pageTableStart, 0)
	cpu.initL2Table(1, l2pageTableStart, 0)

	// set memory region attributes
	write_mair_el3((MemoryRegion << 8) | DeviceRegion)

	// configure translation control register
	var tcr uint32
	bits.SetN(&tcr, TCR_T0SZ, 0x3f, 16)    // memory region size offset 0:5
	bits.SetN(&tcr, TCR_IRGN0, 0b11, 0b01) // inner cacheability (normal, cacheable)
	bits.SetN(&tcr, TCR_ORGN0, 0b11, 0b01) // outer cacheability (normal, cacheable)
	bits.SetN(&tcr, TCR_SH0, 0b11, 0b11)   // inner shareable
	bits.SetN(&tcr, TCR_TG0, 0b11, 0b00)   // 4KB granule
	bits.SetN(&tcr, TCR_PS, 0b111, 0b000)  // 32-bit physical address size (4GB)
	write_tcr_el3(tcr)

	set_ttbr0_el3(l1pageTableStart)
}
