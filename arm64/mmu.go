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
	l1pageTableOffset    = 0x4000
	l2pageTableOffset    = 0x5000
	l3pageTableOffset    = 0x6000
	pageTableArenaOffset = 0x7000

	entriesPerTable = 512
	pageTableSize   = entriesPerTable * 8
)

type mmuMap struct {
	ramStart  uint64
	ramEnd    uint64
	textStart uint64
	textEnd   uint64

	arenaNext uint64
	arenaEnd  uint64
}

type mappingAction uint8

const (
	mappingDevice mappingAction = iota
	mappingMemoryExec
	mappingMemoryXN
	mappingSplit
)

func alignDown(addr uint64, align uint64) uint64 {
	return addr &^ (align - 1)
}

func alignUp(addr uint64, align uint64) uint64 {
	return alignDown(addr+align-1, align)
}

func newMMUMap() (m mmuMap) {
	ramStart, ramEnd := runtime.MemRegion()
	textStart, textEnd := runtime.TextRegion()

	if ramStart&(pageTableSize-1) != 0 || ramEnd&(pageTableSize-1) != 0 {
		panic("RAM region not 4KB aligned")
	}

	m = mmuMap{
		ramStart:  ramStart,
		ramEnd:    ramEnd,
		textStart: alignDown(textStart, pageTableSize),
		textEnd:   alignUp(textEnd, pageTableSize),
	}

	if m.textStart < ramStart || m.textEnd > ramEnd {
		panic("text region outside RAM")
	}

	m.arenaNext = ramStart + pageTableArenaOffset
	m.arenaEnd = m.textStart

	if m.arenaNext >= m.arenaEnd {
		panic("empty early page table space; link binary higher in RAM")
	}

	return
}

func (m *mmuMap) alloc() (addr uint64) {
	addr = m.arenaNext

	if addr+pageTableSize > m.arenaEnd {
		panic("out of early page table space; link binary higher in RAM")
	}

	m.arenaNext += pageTableSize

	return
}

func (m *mmuMap) classifyBlock(addr uint64, end uint64) (action mappingAction) {
	switch {
	case addr >= m.ramStart && end <= m.textEnd:
		// exception vector table
		action = mappingMemoryExec
	case addr >= m.textEnd && end <= m.ramEnd:
		// runtime memory
		action = mappingMemoryXN
	case addr >= m.ramStart && end <= m.ramEnd:
		// text/runtime boundary cross
		action = mappingSplit
	case addr < m.ramEnd && end > m.ramStart:
		// RAM boundary cross
		action = mappingSplit
	default:
		action = mappingDevice
	}

	return
}

func (m *mmuMap) classifyPage(addr uint64) (action mappingAction) {
	end := addr + pageTableSize
	action = m.classifyBlock(addr, end)

	if action == mappingSplit {
		panic("MMU boundary is not 4KB aligned")
	}

	return
}

func writeMapping(page uint64, addr uint64, memoryRegion uint64, deviceRegion uint64, action mappingAction) {
	switch action {
	case mappingMemoryExec:
		reg.Write64(page, addr|memoryRegion)
	case mappingMemoryXN:
		reg.Write64(page, addr|memoryRegion|TTE_EXECUTE_NEVER)
	case mappingDevice:
		reg.Write64(page, addr|deviceRegion|TTE_EXECUTE_NEVER)
	default:
		panic("invalid MMU mapping action")
	}
}

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

	deviceAttributes = 1<<TTE_AF | TTE_OUTER_SH | TTE_AP_00<<TTE_AP | deviceAttributeIndex<<TTE_ATTR
	memoryAttributes = 1<<TTE_AF | TTE_INNER_SH | TTE_AP_00<<TTE_AP | memoryAttributeIndex<<TTE_ATTR
)

// MMU access permissions
//
// ARM Architecture Reference Manual ARMv8,for ARMv8-A architecture profile
// Table D5-25, Data access permissions for stage 1 translations.
const (
	// EL1 or above: read/write, EL0: none
	TTE_AP_00 = 0b00
	// EL1 or above: read/write, EL0: read/write
	TTE_AP_01 = 0b01
	// EL1 or above: read-only, EL0: none
	TTE_AP_10 = 0b10
	// EL1 or above: read-only, EL0: read-only
	TTE_AP_11 = 0b11
)

// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile
// D12.2.103 TCR_EL1, Translation Control Register (EL1).
const (
	TCR_IPS   = 32
	TCR_TBID  = 29
	TCR_EPD1  = 23
	TCR_TG0   = 14
	TCR_SH0   = 12
	TCR_ORGN0 = 10
	TCR_IRGN0 = 8
	TCR_T0SZ  = 0

	// 32-bit or 40-bit intermediate physical address size
	tcr uint64 = 0b010<<TCR_IPS |
		// disable TTBR1_EL1 translation table walks
		0b1<<TCR_EPD1 |
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
func write_mair_el1(val uint64)
func write_tcr_el1(val uint64)
func set_ttbr0_el1(addr uint64)

// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile
// D5.3.1 Translation table level 0, level 1, and level 2 descriptor formats.
func (m *mmuMap) initL1Table(entry int, ttbr uint64, section uint64) {
	n := 30 // 1GB

	memoryRegion := memoryAttributes | TTE_BLOCK
	deviceRegion := deviceAttributes | TTE_BLOCK

	for i := uint64(entry); i < entriesPerTable; i++ {
		page := ttbr + 8*i
		addr := section + (i << n)
		end := addr + (1 << n)
		action := m.classifyBlock(addr, end)

		if action == mappingSplit {
			next := m.alloc()
			reg.Write64(page, next|TTE_TABLE)
			m.initL2Table(0, next, addr)
		} else {
			writeMapping(page, addr, memoryRegion, deviceRegion, action)
		}
	}
}

// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile
// D5.3.1 Translation table level 0, level 1, and level 2 descriptor formats.
func (m *mmuMap) initL2Table(entry int, base uint64, section uint64) {
	n := 21 // 2MB

	memoryRegion := memoryAttributes | TTE_BLOCK
	deviceRegion := deviceAttributes | TTE_BLOCK

	for i := uint64(entry); i < entriesPerTable; i++ {
		page := base + 8*i
		addr := section + (i << n)
		end := addr + (1 << n)
		action := m.classifyBlock(addr, end)

		if action == mappingSplit {
			next := m.alloc()
			reg.Write64(page, next|TTE_TABLE)
			m.initL3Table(0, next, addr)
		} else {
			writeMapping(page, addr, memoryRegion, deviceRegion, action)
		}
	}
}

// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile
// D5.3.2 ARMv8 translation table level 3 descriptor formats.
func (m *mmuMap) initL3Table(entry int, base uint64, section uint64) {
	n := 12 // 4KB

	memoryRegion := memoryAttributes | TTE_PAGE
	deviceRegion := deviceAttributes | TTE_PAGE

	for i := uint64(entry); i < entriesPerTable; i++ {
		page := base + 8*i
		addr := section + (i << n)

		writeMapping(page, addr, memoryRegion, deviceRegion, m.classifyPage(addr))
	}
}

// InitMMU initializes translation tables for all available memory with a flat
// mapping.
//
// The first 4096 bytes (0x00000000 - 0x00001000) are flagged as invalid to
// trap null pointers.
//
// All available memory is marked as non-executable except for the range
// returned by runtime.TextRegion().
func (cpu *CPU) InitMMU() {
	m := newMMUMap()

	l1pageTableStart := m.ramStart + l1pageTableOffset
	l2pageTableStart := m.ramStart + l2pageTableOffset
	l3pageTableStart := m.ramStart + l3pageTableOffset

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
	m.initL1Table(1, l1pageTableStart, 0)
	m.initL2Table(1, l2pageTableStart, 0)
	m.initL3Table(1, l3pageTableStart, 0)

	// set memory region attributes
	//   * attr0: device
	//   * attr1: memory
	write_mair_el1(
		MemoryRegion<<(8*memoryAttributeIndex) |
			DeviceRegion<<(8*deviceAttributeIndex))

	// set translation control register
	write_tcr_el1(tcr)

	// enable MMU
	set_ttbr0_el1(l1pageTableStart)
}
