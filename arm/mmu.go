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

// InitMMU initializes the first-level translation tables for all available
// memory with a flat mapping and privileged attribute flags.
//
// The first 4096 bytes (0x00000000 - 0x00001000) are flagged as invalid to
// trap null pointers, applications that need to make use of this memory space
// must use ConfigureMMU to reconfigure as required.
func (cpu *CPU) InitMMU() {
	ramStart, ramEnd := runtime.MemRegion()

	l1pageTableStart := vecTableStart + l1pageTableOffset
	l2pageTableStart := vecTableStart + l2pageTableOffset

	// First level address translation
	// 9.4, ARM® Cortex™ -A Series Programmer’s Guide

	// The first section is mapped with an L2 entry as we need to map the
	// smallest possible section starting from 0x0 as invalid to trap null
	// pointers.
	firstSection := l2pageTableStart | TTE_PAGE_TABLE
	reg.Write(l1pageTableStart, firstSection)

	for i := uint32(1); i < l1pageTableSize/4; i++ {
		page := l1pageTableStart + 4*i
		pa := i << 20

		if pa >= ramStart && pa < ramEnd {
			reg.Write(page, pa|MemoryRegion)
		} else {
			reg.Write(page, pa|DeviceRegion)
		}
	}

	// Level 2 translation tables
	// 9.5, ARM® Cortex™ -A Series Programmer’s Guide

	memoryRegion := TTE_AP_001<<4 | TTE_CACHEABLE | TTE_BUFFERABLE | TTE_SECTION
	deviceRegion := TTE_AP_001<<4 | TTE_SECTION

	// trap nil pointers by setting the first 4KB section as invalid
	reg.Write(l2pageTableStart, 0)

	for i := uint32(1); i < 256; i++ {
		page := l2pageTableStart + 4*i
		pa := i << 12

		if pa >= ramStart && pa < ramEnd {
			reg.Write(page, pa|memoryRegion)
		} else {
			reg.Write(page, pa|deviceRegion)
		}
	}

	set_ttbr0(l1pageTableStart)
}

// ConfigureMMU (re)configures the first-level translation tables for the
// provided memory range with the argument attribute flags. An alias argument
// greater than zero specifies the physical address corresponding to the start
// argument in case virtual memory is required, otherwise a flat 1:1 mapping is
// set.
func (cpu *CPU) ConfigureMMU(start, end, alias, flags uint32) {
	l1pageTableStart := vecTableStart + l1pageTableOffset

	start = start >> 20
	end = end >> 20
	alias = alias >> 20

	var pa uint32

	for i := start; i < l1pageTableSize/4; i++ {
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
	cpu.FlushTLBs()
}

func (cpu *CPU) updateMMU(start uint32, end uint32, pos int, mask int, val uint32) {
	l1pageTableStart := vecTableStart + l1pageTableOffset

	start = start >> 20
	end = end >> 20

	for i := start; i < l1pageTableSize/4; i++ {
		if i >= end {
			break
		}

		page := l1pageTableStart + 4*i
		reg.SetN(page, pos, mask, val)
	}

	cpu.FlushDataCache()
	cpu.FlushTLBs()
}

// SetAccessPermissions (re)configures the first-level translation tables for
// the provided memory range with the argument domain and access permissions.
func (cpu *CPU) SetAccessPermissions(start, end, ap, domain uint32) {
	cpu.updateMMU(start, end, 5, 0b1101111, (ap<<5)|(domain&0xf))
}

// SetFlags (re)configures the first-level translation tables for the provided
// memory range with the argument attribute flags.
func (cpu *CPU) SetAttributes(start, end, flags uint32) {
	mask := TTE_NS | TTE_SUPERSECTION | TTE_EXECUTE_NEVER | TTE_CACHEABLE |
		TTE_BUFFERABLE | TTE_SECTION | TTE_PAGE_TABLE

	cpu.updateMMU(start, end, 0, int(mask), flags)
}
