// ARM processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm

import (
	"runtime"

	"github.com/usbarmory/tamago/internal/reg"
)

const (
	l1pageTableOffset = 0x4000
	l1pageTableSize   = 0x4000
	l2pageTableOffset = 0xc000
	l2pageTableSize   = 0x4000
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
// (Table B3-8, ARM Architecture Reference Manual ARMv7-A and ARMv7-R edition).
const (
	// PL1: no access   PL0: no access
	TTE_AP_000 uint32 = 0b000
	// PL1: read/write  PL0: no access
	TTE_AP_001 uint32 = 0b001
	// PL1: read/write  PL0: read only
	TTE_AP_010 uint32 = 0b010
	// PL1: read/write  PL0: read/write
	TTE_AP_011 uint32 = 0b011
	// Reserved
	TTE_AP_100 uint32 = 0b100
	// PL1: read only   PL0: no access
	TTE_AP_101 uint32 = 0b101
	// PL1: read only   PL0: read only
	TTE_AP_110 uint32 = 0b110
	// PL1: read only   PL0: read only
	TTE_AP_111 uint32 = 0b111
)

// defined in mmu.s
func flush_tlb()
func set_ttbr0(addr uint32)

// ConfigureMMU (re)configures the first-level translation tables for the
// provided memory range with the passed attribute flags. An alias argument
// greater than zero specifies the physical address corresponding to the start
// argument in case virtual memory is required, otherwise a flat 1:1 mapping is
// set.
func (cpu *CPU) ConfigureMMU(start uint32, end uint32, alias uint32, flags uint32) {
	ramStart, _ := runtime.MemRegion()
	l1pageTableStart := ramStart + l1pageTableOffset

	start = start >> 20
	end = end >> 20
	alias = alias >> 20

	var pa uint32

	for i := uint32(0); i < l1pageTableSize/4; i++ {
		if i < start {
			continue
		}

		if i >= end {
			break
		}

		page := l1pageTableStart + 4*i

		if alias > 0 {
			pa = (alias + i - start) << 20
		} else {
			pa = i << 20
		}

		reg.Write(page, pa|flags)
	}

	cpu.FlushDataCache()
	flush_tlb()
}

// InitMMU initializes the first-level translation tables for all available
// memory with a flat mapping and privileged attribute flags.
//
// The first 4096 bytes (0x00000000 - 0x00001000) are flagged as invalid to
// trap null pointers, applications that need to make use of this memory space
// must use ConfigureMMU to reconfigure as required.
func (cpu *CPU) InitMMU() {
	ramStart, ramEnd := runtime.MemRegion()

	l1pageTableStart := ramStart + l1pageTableOffset
	l2pageTableStart := ramStart + l2pageTableOffset

	// First level address translation
	// 9.4, ARM® Cortex™ -A Series Programmer’s Guide

	memAttr := (TTE_AP_001&0b11)<<10 | TTE_CACHEABLE | TTE_BUFFERABLE | TTE_SECTION
	devAttr := (TTE_AP_001&0b11)<<10 | TTE_SECTION

	// The first section is mapped with an L2 entry as we need to map the
	// smallest possible section starting from 0x0 as invalid to trap null
	// pointers.
	firstSection := l2pageTableStart | TTE_PAGE_TABLE
	reg.Write(l1pageTableStart, firstSection)

	for i := uint32(1); i < l1pageTableSize/4; i++ {
		page := l1pageTableStart + 4*i
		pa := i << 20

		if pa >= ramStart && pa < ramEnd {
			reg.Write(page, pa|memAttr)
		} else {
			reg.Write(page, pa|devAttr)
		}
	}

	// Level 2 translation tables
	// 9.5, ARM® Cortex™ -A Series Programmer’s Guide

	memAttr = (TTE_AP_001&0b11)<<4 | TTE_CACHEABLE | TTE_BUFFERABLE | TTE_SECTION
	devAttr = (TTE_AP_001&0b11)<<4 | TTE_SECTION

	// trap nil pointers by setting the first 4KB section as invalid
	reg.Write(l2pageTableStart, 0)

	for i := uint32(1); i < 256; i++ {
		page := l2pageTableStart + 4*i
		pa := i << 12

		if pa >= ramStart && pa < ramEnd {
			reg.Write(page, pa|memAttr)
		} else {
			reg.Write(page, pa|devAttr)
		}
	}

	set_ttbr0(l1pageTableStart)
}
