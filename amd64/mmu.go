// AMD64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package amd64

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
)

const (
	CR0_WP   = 16
	CR4_LA57 = 12
)

// Memory region attributes
const (
	TTE_P   uint64 = (1 << 0)
	TTE_RW  uint64 = (1 << 1)
	TTE_PCD uint64 = (1 << 4)
	TTE_PS  uint64 = (1 << 7)

	MemoryRegion = TTE_PS | TTE_RW | TTE_P
	DeviceRegion = TTE_PS | TTE_PCD | TTE_RW | TTE_P
)

// Page levels
const (
	PML4 = 4
	PDPT = 3
	PD   = 2
	PT   = 1

	// AMD64 Architecture Programmer’s Manual
	// Figure 5-17. 4-Kbyte Page Translation—Long Mode 4-Level Paging
	indexPML4    = 39
	indexPDPT    = 30
	indexPD      = 21
	indexPT      = 12
	tableEntries = 512
	indexMask    = 0x1ff
	addrMask     = 0x000ffffffffff000

	pageTableSize = tableEntries * 8
)

type mmu struct {
	sync.Mutex

	arenaNext uint64
	arenaEnd  uint64
}

func alignDown(addr uint64, align uint64) uint64 {
	return addr &^ (align - 1)
}

func alignUp(addr uint64, align uint64) uint64 {
	return alignDown(addr+align-1, align)
}

func (m *mmu) alloc() (addr uint64) {
	if m.arenaEnd == 0 {
		ramStart, _ := runtime.MemRegion()
		textStart, _ := runtime.TextRegion()

		m.arenaNext = alignUp(ramStart, pageTableSize)
		m.arenaEnd = alignDown(textStart, pageTableSize)
	}

	if m.arenaNext+pageTableSize > m.arenaEnd {
		panic("out of page table space; link binary higher in RAM")
	}

	addr = m.arenaNext
	m.arenaNext += pageTableSize

	for i := uint64(0); i < tableEntries; i++ {
		reg.Write64(addr+8*i, 0)
	}

	return
}

// defined in mmu.s
func read_cr0() uint64
func write_cr0(val uint64)
func read_cr3() uint64
func read_cr4() uint64
func set_ext_pdpt(index int, flags uint64)
func flush_tlb()

// SetWriteProtect configures the Write Protect (WP) bit in Control Register 0
// (CR0).
func (cpu *CPU) SetWriteProtect(enable bool) {
	cr0 := read_cr0()
	bits.SetTo64(&cr0, CR0_WP, enable)
	write_cr0(cr0)
}

// FindPTE returns the Page Table Entry (PTE) offset, level and base address
// for a given address. A non-zero address bit length can override the default
// value to support implementation specific limitations (e.g. C-Bit).
func (cpu *CPU) FindPTE(addr uint64, bitLen int) (pte uint64, level int, page uint64) {
	indices := [4]uint64{
		(addr >> indexPML4) & indexMask,
		(addr >> indexPDPT) & indexMask,
		(addr >> indexPD) & indexMask,
		(addr >> indexPT) & indexMask,
	}

	mask := uint64(addrMask)

	if bitLen > 0 {
		mask &= uint64(1<<bitLen) - 1
	}

	pml4 := read_cr3()
	tableAddr := pml4 & mask

	for i := range 4 {
		level = 4 - i
		off := tableAddr + indices[i]*8
		entry := reg.Read64(off)

		if entry&TTE_P == 0 {
			return 0, level, entry & mask
		}

		if (level == PDPT || level == PD) && (entry&TTE_PS != 0) {
			return off, level, entry & mask
		}

		if level == PT {
			return off, PT, entry & mask
		}

		tableAddr = entry & mask
	}

	return 0, 0, 0
}

// SetEncryptedBit (re)configures the page encryption attribute bit (C-bit) for
// a given memory range, an error is raised if the argument range spawns across
// multiple translation levels or is not page aligned.
func (cpu *CPU) SetEncryptedBit(start uint64, end uint64, cbit int, private bool) (err error) {
	startPTE, startLevel, startPage := cpu.FindPTE(start, cbit)
	endPTE, endLevel, _ := cpu.FindPTE(end, cbit)

	if startLevel != endLevel {
		return fmt.Errorf("changing C-bit on multiple translation levels is unsupported")
	}

	if start != startPage {
		return fmt.Errorf("start address (%#x) does not match PTE base address (%#x)", start, startPage)
	}

	cpu.SetWriteProtect(false)
	defer cpu.SetWriteProtect(true)

	for pte := startPTE; pte < endPTE; pte += 8 {
		reg.SetTo64(pte, cbit, private)
	}

	return
}

// ConfigurePDPT (re)configures the third-level translation tables (PDPT) for
// the provided extended memory range (i.e. above 4 GiB, range must be 1 GiB
// aligned) with the argument attribute flags, a flat 1:1 mapping is set.
func (cpu *CPU) ConfigurePDPT(start, end, flags uint64) {
	var index int

	// ensure extended memory range
	if start < (1 << 32) {
		return
	}

	// force lowest level of the page-translation hierarchy
	flags |= TTE_PS

	for base := start; base < end; base += 1 << 30 {
		if index = int(base >> 30); index >= tableEntries {
			break
		}

		set_ext_pdpt(index, flags)
	}

	flush_tlb()
}

func tableIndex(addr uint64, shift int) uint64 {
	return (addr >> shift) & indexMask
}

// pdEntry returns the Page Directory (2 MB) entry address for the argument
// virtual address, allocating missing tables and splitting a 1 GB page into
// 2 MB pages with the same attributes if needed.
func (cpu *CPU) pdEntry(addr uint64) uint64 {
	pml4e := (read_cr3() & addrMask) + 8*tableIndex(addr, indexPML4)

	if reg.Read64(pml4e)&TTE_P == 0 {
		reg.Write64(pml4e, cpu.mmu.alloc()|TTE_RW|TTE_P)
	}

	pdpte := (reg.Read64(pml4e) & addrMask) + 8*tableIndex(addr, indexPDPT)
	tte := reg.Read64(pdpte)

	switch {
	case tte&TTE_P == 0:
		reg.Write64(pdpte, cpu.mmu.alloc()|TTE_RW|TTE_P)
	case tte&TTE_PS != 0:
		next := cpu.mmu.alloc()

		// the large page attributes (including PAT) keep their position
		for i := uint64(0); i < tableEntries; i++ {
			reg.Write64(next+8*i, tte+i<<indexPD)
		}

		reg.Write64(pdpte, next|TTE_RW|TTE_P)
	}

	return (reg.Read64(pdpte) & addrMask) + 8*tableIndex(addr, indexPD)
}

// ConfigureMMU (re)configures the page tables for the provided memory range
// with the argument attribute flags. An alias argument greater than zero
// specifies the physical address corresponding to the start argument in case
// virtual memory is required, otherwise a flat 1:1 mapping is set.
func (cpu *CPU) ConfigureMMU(start, end, alias, flags uint64) {
	const size = 1 << indexPD

	if cr4 := read_cr4(); bits.Get64(&cr4, CR4_LA57) {
		panic("5-level paging is not supported")
	}

	for _, addr := range []uint64{start, end - 1} {
		if top := addr >> 47; top != 0 && top != 0x1ffff {
			panic("non-canonical address")
		}
	}

	if alias > 0 && (alias-start)%size != 0 {
		panic("alias and start offset differ within 2 MB")
	}

	start, alias = alignDown(start, size), alignDown(alias, size)
	end = alignUp(end, size)

	cpu.mmu.Lock()
	defer cpu.mmu.Unlock()

	// page tables might be write protected
	cr0 := read_cr0()
	cpu.SetWriteProtect(false)
	defer write_cr0(cr0)

	for addr := start; addr < end; addr += size {
		pa := addr

		if alias > 0 {
			pa = alias + (addr - start)
		}

		pde := cpu.pdEntry(addr)

		if tte := reg.Read64(pde); tte&TTE_P != 0 && tte&TTE_PS == 0 {
			// update existing 4 KB pages, bit 7 is PAT instead of PS
			pt := tte & addrMask

			for i := uint64(0); i < tableEntries; i++ {
				reg.Write64(pt+8*i, (pa+i<<indexPT)|flags&^TTE_PS)
			}

			continue
		}

		reg.Write64(pde, pa|flags|TTE_PS)
	}

	flush_tlb()
}
