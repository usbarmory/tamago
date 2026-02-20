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

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
)

const CR0_WP = 16

// Memory region attributes
const (
	TTE_P  uint64 = (1 << 0)
	TTE_PS uint64 = (1 << 7)
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
)

// defined in mmu.s
func read_cr0() uint64
func write_cr0(val uint64)
func read_cr3() uint64

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
