// ARM processor support
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// +build tamago,arm staticcheck

package arm

import (
	"runtime"

	"github.com/f-secure-foundry/tamago/internal/reg"
)

const (
	l1pageTableOffset = 0x4000 // 16 kB
	l1pageTableSize   = 0x4000 // 16 kB
)

// Memory region attributes
// Table B3-10 ARM Architecture Reference Manual ARMv7-A and ARMv7-R edition
const (
	TTE_SECTION_1MB   uint32 = 0x2
	TTE_SECTION_16MB  uint32 = 0x40002
	TTE_EXECUTE_NEVER uint32 = 0x10
	TTE_CACHEABLE     uint32 = 0x8
	TTE_BUFFERABLE    uint32 = 0x4
)

// MMU access permissions
// Table B3-8 ARM Architecture Reference Manual ARMv7-A and ARMv7-R edition
const (
	// PL1: no access   PL0: no access
	TTE_AP_000 uint32 = 0b000000 << 10
	// PL1: read/write  PL0: no access
	TTE_AP_001 uint32 = 0b000001 << 10
	// PL1: read/write  PL0: read only
	TTE_AP_010 uint32 = 0b000010 << 10
	// PL1: read/write  PL0: read/write
	TTE_AP_011 uint32 = 0b000011 << 10
	// Reserved
	TTE_AP_100 uint32 = 0b100000 << 10
	// PL1: read only   PL0: no access
	TTE_AP_101 uint32 = 0b100001 << 10
	// PL1: read only   PL0: read only
	TTE_AP_110 uint32 = 0b100010 << 10
	// PL1: read only   PL0: read only
	TTE_AP_111 uint32 = 0b100011 << 10
)

// defined in mmu.s
func set_ttbr0(addr uint32)

// ConfigureMMU (re)configures the first-level translation tables for the
// provided memory range with the passed attribute flags.
func (cpu *CPU) ConfigureMMU(start uint32, end uint32, flags uint32) {
	ramStart, _ := runtime.MemRegion()
	l1pageTableStart := ramStart + l1pageTableOffset

	start = start >> 20
	end = end >> 20

	for i := uint32(1); i < l1pageTableSize/4; i++ {
		page := l1pageTableStart + 4*i
		pa := i << 20

		if i < start {
			continue
		}

		if i >= end {
			break
		}

		reg.Write(page, pa|flags)
	}

	set_ttbr0(l1pageTableStart)
}

// InitMMU initializes the first-level translation tables for all available
// memory with a flat mapping and privileged attribute flags.
func (cpu *CPU) InitMMU() {
	start, end := runtime.MemRegion()
	l1pageTableStart := start + l1pageTableOffset

	memAttr := uint32(TTE_AP_001 | TTE_CACHEABLE | TTE_BUFFERABLE | TTE_SECTION_1MB)
	devAttr := uint32(TTE_AP_001 | TTE_SECTION_1MB)

	start = start >> 20
	end = end >> 20

	// skip page zero to trap null pointers
	reg.Write(l1pageTableStart, 0)

	for i := uint32(1); i < l1pageTableSize/4; i++ {
		page := l1pageTableStart + 4*i
		pa := i << 20

		if i >= start && i < end {
			reg.Write(page, pa|memAttr)
		} else {
			reg.Write(page, pa|devAttr)
		}
	}

	set_ttbr0(l1pageTableStart)
}
