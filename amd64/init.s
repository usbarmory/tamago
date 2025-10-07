// x86-64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkcpuinit

#include "amd64.h"
#include "go_asm.h"
#include "textflag.h"

// Global Descriptor Table
DATA	·gdt<>+0x00(SB)/8, $0x0000000000000000	// null descriptor
DATA	·gdt<>+0x08(SB)/8, $0x00209a0000000000	// code descriptor (x/r)
DATA	·gdt<>+0x10(SB)/8, $0x0000920000000000	// data descriptor (r/w)
GLOBL	·gdt<>(SB),RODATA,$24

DATA	·gdtptr+0x00(SB)/2, $(3*8-1)		// GDT Limit
DATA	·gdtptr+0x02(SB)/8, $·gdt<>(SB)		// GDT Base Address
GLOBL	·gdtptr(SB),RODATA,$(2+8)

TEXT cpuinit(SB),NOSPLIT|NOFRAME,$0
	// disable interrupts
	CLI

	// we might not have a valid stack pointer for CALLs
	MOVL	$PML4T, SP

	// Set up paging
	//
	// AMD64 Architecture Programmer’s Manual
	// Volume 2 - 5.3 Long-Mode Page Translation
	//
	// Intel® 64 and IA-32 Architectures Software Developer’s Manual
	// Volume 3A - 4.5 IA-32E PAGING

	// clear tables
	XORL	AX, AX		// value
	MOVL	$PML4T, DI	// to
	MOVL	$0x4000, CX	// n
	MOVL	DI, CR3
	REP;	STOSB

	// PML4T[0] = PDPT
	MOVL	$PML4T, DI
	MOVL	$(PDPT | 1<<1 | 1<<0), (DI)			// set R/W, P

	// PDPT[0] = PDT
	MOVL	$PDPT, DI
	MOVL	$(PDT | 1<<1 | 1<<0), (DI)			// set R/W, P

	// PDPT[1]: 0xc0000000 - 0xffffffff (1GB) uncacheable physical page (1GB PDPE)
	ADDL	$8, DI
	MOVL	$(1<<30 | 1<<7 | 1<<1 | 1<<0), (DI)		// set PS, R/W, P

	// PDPT[2]: 0x80000000 - 0xbfffffff (1GB) cacheable physical page (1GB PDPE)
	ADDL	$8, DI
	MOVL	$(2<<30 | 1<<7 | 1<<1 | 1<<0), (DI)		// set PS, R/W, P

	// PDPT[3]: 0x40000000 - 0x7fffffff (1GB) cacheable physical page (1GB PDPE)
	ADDL	$8, DI
	MOVL	$(3<<30 | 1<<7 | 1<<4 | 1<<1 | 1<<0), (DI)	// set PS, PCD, R/W, P

	// PDT[..]: 0x00000000 - 0x3fffffff (1GB) cacheable physical page (2MB PDTEs)
	MOVL	$PDT, DI
	MOVL	$0, AX
add_pdt_entries:
	CMPL	AX, $(1 << 30)
	JAE	check_long_mode

	ORL	$(1<<7 | 1<<1 | 1<<0), AX			// set PS, R/W, P
	MOVL	AX, (DI)

	ADDL	$(2<<20), AX
	ADDL	$8, DI
	JMP	add_pdt_entries

check_long_mode:
	MOVL	CR4, AX
	ANDL	$(1<<7 | 1<<5), AX	// get CR4.(PGE|PAE)
	JBE	enable_long_mode

	JMP	·reload_gdt<>(SB)

enable_long_mode:
	MOVL	$(1<<7 | 1<<5), AX	// set CR4.(PGE|PAE)
	MOVL	AX, CR4

	MOVL	$MSR_EFER, CX
	RDMSR
	ORL	$(1<<8), AX		// set MSR_EFER.LME
	WRMSR

	MOVL	CR0, AX
	ORL	$(1<<31 | 1<<0), AX	// set CR0.(PG|PE)
	MOVL	AX, CR0

	// set Global Descriptor Table
	CALL	·getPC<>(SB)
	MOVL	$·gdtptr(SB), BX	// 32-bit mode: only PC offset is copied
	ADDL	$6, AX
	ADDL	BX, AX
	LGDT	(AX)

	// set far return target
	CALL	·getPC<>(SB)
	MOVL	$·start<>(SB), BX	// 32-bit mode: only PC offset is copied
	ADDL	$6, AX
	ADDL	BX, AX

	// jump to target in long mode
	PUSHQ	$0x08
	PUSHQ	AX
	RETFQ

TEXT ·reload_gdt<>(SB),NOSPLIT|NOFRAME,$0
	MOVQ	$·gdtptr(SB), AX
	LGDT	(AX)

	MOVQ	$·start<>(SB), AX

	PUSHQ	$0x08
	PUSHQ	AX
	RETFQ

TEXT ·start<>(SB),NOSPLIT|NOFRAME,$0
	// enable SSE
	CALL	sse_enable(SB)

	// PDT[0] = PT
	MOVL	$PDT, DI
	MOVL	$(PT | 1<<1 | 1<<0), (DI)			// set R/W, P

	// PT[0]:  0x00000000 - 0x00001000 inaccessible (zero page)
	MOVL	$PT, DI
	ANDL	$(1<<1 | 1<<0), (DI)				// clear R/W, P

	// PT[..]: 0x00001000 - 0x001fffff cacheable physical page (4KB PTEs)
	ADDL	$8, DI
	MOVL	$(4<<10), AX
add_pt_entries:
	CMPL	AX, $(1 << 21)
	JAE	add_ext_entries

	ORL	$(1<<1 | 1<<0), AX				// set R/W, P
	MOVL	AX, (DI)

	ADDL	$(4<<10), AX
	ADDL	$8, DI
	JMP	add_pt_entries

add_ext_entries:
	// add extended Long-Mode Page Translation PDT (1GB) entries.
	MOVL	$PDPT, DI

	// PDPT[4]: 0x100000000 - 0x13fffffff (1GB) uncacheable physical page (1GB PDPE)
	ADDL	$(8*4), DI
	MOVQ	$(4<<30 | 1<<7 | 1<<4 | 1<<1 | 1<<0), AX	// set PS, PCD, R/W, P
	MOVQ	AX, (DI)

	// PDPT[5]: 0x140000000 - 0x17fffffff (1GB) uncacheable physical page (1GB PDPE)
	ADDL	$8, DI
	MOVQ	$(5<<30 | 1<<7 | 1<<4 | 1<<1 | 1<<0), AX	// set PS, PCD, R/W, P
	MOVQ	AX, (DI)

	// flush TLBs
	MOVL	$PML4T, DI
	MOVL	DI, CR3

	STI	// required for ·wfi
	JMP	_rt0_tamago_start(SB)

TEXT ·getPC<>(SB),NOSPLIT|NOFRAME,$0
	POPQ	AX
	CALL	AX
