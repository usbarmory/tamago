// ARM processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm64

import (
	"runtime"

	"github.com/usbarmory/tamago/internal/reg"
)

const (
	l1pageTableOffset = 0x4000
	l1pageTableSize   = 4096

	l2pageTableOffset = 0xc000
	l2pageTableSize   = 256
)

// Memory region attributes
// (Table B3-10, ARM Architecture Reference Manual ARMv7-A and ARMv7-R edition).
const (
	TTE_PAGE_TABLE    uint32 = (1 << 0)
	TTE_SECTION       uint32 = (1 << 1)
	TTE_BUFFERABLE    uint32 = (1 << 2)
	TTE_CACHEABLE     uint32 = (1 << 3)
	TTE_EXECUTE_NEVER uint32 = (1 << 4)
	TTE_SUPERSECTION  uint32 = (1 << 18) | (1 << 1)
	TTE_NS            uint32 = (1 << 19)
)

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

const (
	MemoryRegion = TTE_AP_001<<10 | TTE_CACHEABLE | TTE_BUFFERABLE | TTE_SECTION
	DeviceRegion = TTE_AP_001<<10 | TTE_SECTION
)

// defined in mmu.s
func flush_tlb()
func set_ttbr0(addr uint32)

// First level address translation
// 9.4, ARM® Cortex™ -A Series Programmer’s Guide
func (cpu *CPU) initL1Table(entry int, ttbr uint32, section uint32) {
	ramStart, ramEnd := runtime.MemRegion()
	_, textEnd := runtime.TextRegion()

	for i := uint32(entry); i < l1pageTableSize; i++ {
		page := ttbr + 4*i
		addr := section + (i << 20)

		switch {
		case addr < textEnd && (addr+(1<<20)) > textEnd:
			// skip first L2 table, reserved to trap null pointers
			l2pageTableStart := cpu.vbar + l2pageTableOffset
			base := l2pageTableStart + l2pageTableSize*4

			// use L2 table to end non-executable boundary
			// precisely at textStart
			reg.Write(page, base|TTE_PAGE_TABLE)
			cpu.initL2Table(0, base, addr)
		case addr >= ramStart && addr < textEnd:
			reg.Write(page, addr|MemoryRegion)
		case addr >= ramStart && addr < ramEnd:
			reg.Write(page, addr|MemoryRegion|TTE_EXECUTE_NEVER)
		default:
			reg.Write(page, addr|DeviceRegion|TTE_EXECUTE_NEVER)
		}
	}
}

// Level 2 translation tables
// 9.5, ARM® Cortex™ -A Series Programmer’s Guide
func (cpu *CPU) initL2Table(entry int, base uint32, section uint32) {
	ramStart, ramEnd := runtime.MemRegion()
	_, textEnd := runtime.TextRegion()

	memoryRegion := TTE_AP_001<<4 | TTE_CACHEABLE | TTE_BUFFERABLE | TTE_SECTION
	deviceRegion := TTE_AP_001<<4 | TTE_SECTION

	for i := uint32(entry); i < l2pageTableSize; i++ {
		page := base + 4*i
		addr := section + (i << 12)

		switch {
		case addr >= ramStart && addr < textEnd:
			reg.Write(page, addr|memoryRegion)
		case addr >= ramStart && addr < ramEnd:
			reg.Write(page, addr|memoryRegion|TTE_EXECUTE_NEVER)
		default:
			reg.Write(page, addr|deviceRegion|TTE_EXECUTE_NEVER)
		}
	}
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
	l1pageTableStart := cpu.vbar + l1pageTableOffset
	l2pageTableStart := cpu.vbar + l2pageTableOffset

	// Map the first L1 entry to an L2 table to trap null pointers within
	// the smallest possible section (4KB starting from 0x00000000).
	firstSection := l2pageTableStart | TTE_PAGE_TABLE
	reg.Write(l1pageTableStart, firstSection)

	// set first L2 entry as invalid
	reg.Write(l2pageTableStart, 0)

	// set remaining entries with flat mapping
	cpu.initL1Table(1, l1pageTableStart, 0)
	cpu.initL2Table(1, l2pageTableStart, 0)

	set_ttbr0(l1pageTableStart)
}
