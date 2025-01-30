// x86-64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

#define MSR_EFER 0xc0000080

#define PML4T 0x9000	// Page Map Level 4 Table       (512GB entries)
#define PDPT  0xa000	// Page Directory Pointer Table   (1GB entries)
#define PDT   0xb000	// Page Directory Table           (2MB entries)
#define PT    0xc000	// Page Table                     (4kB entries)

// Global Descriptor Table
DATA	gdt<>+0x00(SB)/8, $0x0000000000000000	// null descriptor
DATA	gdt<>+0x08(SB)/8, $0x00209a0000000000	// code descriptor (x/r)
DATA	gdt<>+0x10(SB)/8, $0x0000920000000000	// data descriptor (r/w)
GLOBL	gdt<>(SB),RODATA,$24

DATA	gdtptr<>+0x00(SB)/2, $(3*8-1)		// GDT Limit
DATA	gdtptr<>+0x02(SB)/8, $gdt<>(SB)		// GDT Base Address
GLOBL	gdtptr<>(SB),RODATA,$(2+8)

TEXT cpuinit(SB),NOSPLIT|NOFRAME,$0
	// Disable interrupts
	CLI

	// Set up paging
	//
	// AMD64 Architecture Programmer’s Manual
	// Volume 2 - 5.3 Long-Mode Page Translation
	//
	// Intel® 64 and IA-32 Architectures Software Developer’s Manual
	// Volume 3A - 4.5 IA-32E PAGING

	// Clear tables
	XORL	AX, AX		// value
	MOVL	$PML4T, DI	// to
	MOVL	$0x5000, CX	// n
	MOVL	DI, CR3
	REP;	STOSB

	// PML4T[0] = PDPT
	MOVL	$PML4T, DI
	MOVL	$(PDPT | 1<<1 | 1<<0), (DI)	// set R/W, P

	// Configure Long-Mode Page Translation as follows:
	//   0x00000000 - 0x3fffffff (1GB) cacheable   physical page
	//   0x40000000 - 0x7fffffff (1GB) cacheable   physical page
	//   0x80000000 - 0xbfffffff (1GB) cacheable   physical page
	//   0xc0000000 - 0xffffffff (1GB) uncacheable physical page
	MOVL	$PDPT, DI
	MOVQ	$(0<<30 | 1<<7 | 1<<1 | 1<<0), (DI)		// set PS, R/W, P
	ADDL	$8, DI
	MOVQ	$(1<<30 | 1<<7 | 1<<1 | 1<<0), (DI)		// set PS, R/W, P
	ADDL	$8, DI
	MOVQ	$(2<<30 | 1<<7 | 1<<1 | 1<<0), (DI)		// set PS, R/W, P
	ADDL	$8, DI
	MOVQ	$(3<<30 | 1<<7 | 1<<4 | 1<<1 | 1<<0), (DI)	// set PS, PCD, R/W, P

	MOVL	CR4, AX
	ANDL	$(1<<7 | 1 << 5), AX	// get CR4.(PGE|PAE)
	JBE	enable_long_mode

	JMP	·start<>(SB)

enable_long_mode:
	// Enter long mode

	MOVL	$(1<<7 | 1<<5), AX	// set CR4.(PGE|PAE)
	MOVL	AX, CR4

	MOVL	$MSR_EFER, CX
	RDMSR
	ORL	$(1<<8), AX		// set MSR_EFER.LME
	WRMSR

	MOVL	CR0, AX
	ORL	$(1<<31 | 1<<0), AX	// set CR0.(PG|PE)
	MOVL	AX, CR0

	// Set Global Descriptor Table

	CALL	·getPC<>(SB)
	MOVL	$gdtptr<>(SB), BX	// 32-bit mode: only PC offset is copied
	ADDL	$6, AX
	ADDL	BX, AX
	LGDT	(AX)

	CALL	·getPC<>(SB)
	MOVL	$·start<>(SB), BX	// 32-bit mode: only PC offset is copied
	ADDL	$6, AX
	ADDL	BX, AX

	PUSHQ	$0x08
	PUSHQ	AX
	RETFQ

TEXT ·start<>(SB),NOSPLIT|NOFRAME,$0
	MOVL	CR0, AX
	MOVL	CR4, BX

	// Enable SSE
	ANDL	$~(1<<2), AX		// clear CR0.EM
	ORL	$(1<<1), AX		//   set CR0.MP
	ORL	$(1<<10 | 1<<9), BX	//   set CR4.(OSXMMEXCPT|OSFXSR)

	MOVL	AX, CR0
	MOVL	BX, CR4

	JMP	_rt0_tamago_start(SB)

TEXT ·getPC<>(SB),NOSPLIT|NOFRAME,$0
	POPQ	AX
	CALL	AX
