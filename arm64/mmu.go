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

	"github.com/usbarmory/tamago/internal/reg"
)

const (
	l1pageTableOffset = 0x4000
	l1pageTableSize   = 512

	l2pageTableOffset = 0x5000
	l2pageTableSize   = 512

	// another L2 table is appended at 0x6000

	l3pageTableOffset = 0x7000
	l3pageTableSize   = 512
)

// Memory region attributes
//
// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile
// G5.7.
const (
	TTE_XN   = 53
	TTE_AF   = 10
	TTE_SH   = 8
	TTE_AP   = 6
	TTE_ATTR = 2
	TTE_DESC = 0

	TTE_BLOCK         uint64 = (0b01 << TTE_DESC)
	TTE_TABLE         uint64 = (0b11 << TTE_DESC)
	TTE_PAGE          uint64 = (0b11 << TTE_DESC)
	TTE_NON_SH        uint64 = (0b00 << TTE_SH)
	TTE_OUTER_SH      uint64 = (0b10 << TTE_SH)
	TTE_INNER_SH      uint64 = (0b11 << TTE_SH)
	TTE_EXECUTE_NEVER uint64 = (0b11 << TTE_XN)

	// Device-nGnRnE
	DeviceRegion uint64 = 0b00000000
	// Normal, Inner/Outer WB/WA/RA
	MemoryRegion uint64 = 0b11111111

	deviceAttributeIndex = 0
	memoryAttributeIndex = 1

	deviceAttributes = 1<<TTE_AF | TTE_OUTER_SH | TTE_AP_001<<TTE_AP | deviceAttributeIndex<<TTE_ATTR
	memoryAttributes = 1<<TTE_AF | TTE_INNER_SH | TTE_AP_001<<TTE_AP | memoryAttributeIndex<<TTE_ATTR
)

// MMU access permissions
//
// ARM Architecture Reference Manual ARMv8,for ARMv8-A architecture profile
// Table G5-9, ARM Architecture Reference Manual ARMv8.
const (
	// PL1: no access   Unprivileged: no access
	TTE_AP_000 uint64 = 0b000
	// PL1: read/write  Unprivileged: no access
	TTE_AP_001 uint64 = 0b001
)

// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile
// D12.2.105 TCR_EL3, Translation Control Register (EL3).
const (
	TCR_TBID  = 29
	TCR_PS    = 16
	TCR_TG0   = 14
	TCR_SH0   = 12
	TCR_ORGN0 = 10
	TCR_IRGN0 = 8
	TCR_T0SZ  = 0

	tcr uint64 = 0b0<<TCR_TBID |
		// 48-bit physical address size
		0b101<<TCR_PS |
		// 4KB granule
		0b00<<TCR_TG0 |
		// inner shareable
		0b11<<TCR_SH0 |
		// outer cacheability (normal, cacheable)
		0b01<<TCR_ORGN0 |
		// inner cacheability (normal, cacheable)
		0b01<<TCR_IRGN0 |
		// memory region size offset 0:5 (39-bits VA space)
		25<<TCR_T0SZ
)

// defined in mmu.s
func flush_tlb()
func write_mair_el3(val uint64)
func write_tcr_el3(val uint64)
func set_ttbr0_el3(addr uint64)

// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile
// D5.3.1 Translation table level 0, level 1, and level 2 descriptor formats.
func (cpu *CPU) initL1Table(entry int, ttbr uint64, section uint64) {
	n := 30 // 1GB

	ramStart, ramEnd := runtime.MemRegion()
	_, textEnd := runtime.TextRegion()

	memoryRegion := memoryAttributes | TTE_BLOCK
	deviceRegion := deviceAttributes | TTE_BLOCK

	for i := uint64(entry); i < l1pageTableSize; i++ {
		page := ttbr + 8*i
		addr := section + (i << n)

		switch {
		case addr < textEnd && (addr+(1<<n)) > textEnd:
			// skip first L2 table, pointing to L3
			l2pageTableStart := ramStart + l2pageTableOffset
			base := l2pageTableStart + l2pageTableSize*8

			// use L2 table to end non-executable boundary
			// precisely at textStart
			reg.Write64(page, base|TTE_TABLE)
			cpu.initL2Table(0, base, addr)
		case addr >= ramStart && addr < textEnd:
			reg.Write64(page, addr|memoryRegion)
		case addr >= ramStart && addr < ramEnd:
			reg.Write64(page, addr|memoryRegion|TTE_EXECUTE_NEVER)
		default:
			reg.Write64(page, addr|deviceRegion|TTE_EXECUTE_NEVER)
		}
	}
}

// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile
// D5.3.1 Translation table level 0, level 1, and level 2 descriptor formats.
func (cpu *CPU) initL2Table(entry int, base uint64, section uint64) {
	n := 21 // 2MB

	ramStart, ramEnd := runtime.MemRegion()
	_, textEnd := runtime.TextRegion()

	memoryRegion := memoryAttributes | TTE_BLOCK
	deviceRegion := deviceAttributes | TTE_BLOCK

	for i := uint64(entry); i < l2pageTableSize; i++ {
		page := base + 8*i
		addr := section + (i << n)

		switch {
		case addr < textEnd && (addr+(1<<n)) > textEnd:
			// skip first L3 table, reserved to trap null pointers
			l3pageTableStart := ramStart + l3pageTableOffset
			base := l3pageTableStart + l3pageTableSize*8

			// use L3 table to end non-executable boundary
			// precisely at textStart
			reg.Write64(page, base|TTE_TABLE)
			cpu.initL3Table(0, base, addr)
		case addr >= ramStart && addr < textEnd:
			reg.Write64(page, addr|memoryRegion)
		case addr >= ramStart && addr < ramEnd:
			reg.Write64(page, addr|memoryRegion|TTE_EXECUTE_NEVER)
		default:
			reg.Write64(page, addr|deviceRegion|TTE_EXECUTE_NEVER)
		}
	}
}

// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile
// D5.3.2 ARMv8 translation table level 3 descriptor formats.
func (cpu *CPU) initL3Table(entry int, base uint64, section uint64) {
	n := 12 // 4KB

	ramStart, ramEnd := runtime.MemRegion()
	_, textEnd := runtime.TextRegion()

	memoryRegion := memoryAttributes | TTE_PAGE
	deviceRegion := deviceAttributes | TTE_PAGE

	for i := uint64(entry); i < l3pageTableSize; i++ {
		page := base + 8*i
		addr := section + (i << n)

		switch {
		case addr >= ramStart && addr < textEnd:
			reg.Write64(page, addr|memoryRegion)
		case addr >= ramStart && addr < ramEnd:
			reg.Write64(page, addr|memoryRegion|TTE_EXECUTE_NEVER)
		default:
			reg.Write64(page, addr|deviceRegion|TTE_EXECUTE_NEVER)
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
	ramStart, _ := runtime.MemRegion()

	l1pageTableStart := ramStart + l1pageTableOffset
	l2pageTableStart := ramStart + l2pageTableOffset
	l3pageTableStart := ramStart + l3pageTableOffset

	// Map the first L1 entry to an L2 table.
	tte := l2pageTableStart | TTE_TABLE
	reg.Write64(l1pageTableStart, tte)

	// Map the first L2 entry to an L3 table to trap null pointers within
	// the smallest possible section (4KB starting from 0x00000000).
	tte = l3pageTableStart | TTE_TABLE
	reg.Write64(l2pageTableStart, tte)

	// set first L3 entry as invalid
	reg.Write64(l3pageTableStart, 0)

	// set remaining entries with flat mapping
	cpu.initL1Table(1, l1pageTableStart, 0)
	cpu.initL2Table(1, l2pageTableStart, 0)
	cpu.initL3Table(1, l3pageTableStart, 0)

	// set memory region attributes
	//   * attr0: device
	//   * attr1: memory
	write_mair_el3(
		MemoryRegion<<(8*memoryAttributeIndex) |
			DeviceRegion<<(8*deviceAttributeIndex))

	// set translation control register
	write_tcr_el3(tcr)

	// enable MMU
	set_ttbr0_el3(l1pageTableStart)
}
