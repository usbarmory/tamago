// x86-64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

#define MSR_EFER 0xc0000080

#define PML4T 0x1000 // Page Map Level 4 Table       (512GB entries)
#define PDPT  0x2000 // Page Directory Pointer Table   (1GB entries)
#define PDT   0x3000 // Page Directory Table           (2MB entries)
#define PT    0x4000 // Page Table                     (4kB entries)

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

	// PDPT[0] = PDT
	MOVL	$PDPT, DI
	MOVL	$(PDT | 1<<1 | 1<<0), (DI)	// set R/W, P

	// PDT[0] = PT
	MOVL	$PDT, DI
	MOVL	$(PT | 1<<1 | 1<<0), (DI)	// set R/W, P

	MOVL	$PT, DI
	MOVL	$(1<<1 | 1<<0), AX		// set R/W, P
build_table:
	MOVL	AX, (DI)
	ADDL	$0x1000, AX
	ADDL	$8, DI
	CMPL	AX, $0x200000
	JB	build_table

	// Enter long mode

	MOVL	$(1<<7 | 1<<5), AX		// set PGE, PAE
	MOVL	AX, CR4

	MOVL	$MSR_EFER, CX
	RDMSR
	ORL	$(1<<8), AX			// set LME
	WRMSR

	MOVL	CR0, BX
	ORL	$(1<<31 | 1<<0), BX		// set PG, PE
	MOVL	BX, CR0

	// Hello World (debug)
	MOVL	$0xb8000, DI
	MOVQ	$0x1F6C1F6C1F651F48, AX
	MOVQ	AX, (DI)
	MOVQ	$0x1F6F1F571F201F6F, AX
	ADDQ	AX, 1(DI)
	MOVQ	$0x1F211F641F6C1F72, AX
	MOVQ	AX, 2(DI)

	JMP	_rt0_tamago_start(SB)
