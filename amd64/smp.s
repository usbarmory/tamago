// x86-64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

#define MSR_EFER 0xc0000080

// Page Map Level 4 Table (see init.s)
#define PML4T 0x9000

// These legacy prefixes are required to ensure valid Go assembly
// interpretation under 16-bit Real Mode.
#define DATA32 BYTE $0x66	// 32-bit operand size override prefix
#define ADDR32 BYTE $0x67	// 32-bit address size override prefix

// apinit is invoked under 16-bit Real Mode to initialize an Application
// Processor (AP) in SMP operation.
TEXT ·apinit<>(SB),NOSPLIT|NOFRAME,$0
	// disable interrupts
	CLI

	// we might not have a valid stack pointer for CALLs
	DATA32
	MOVL	$PML4T, SP

	// BSP paging setup (see init.s)
	MOVL	SP, CR3

	BYTE	$0x8c; BYTE	$0xc8		// mov %cs,%ax
	BYTE	$0x8e; BYTE	$0xd8		// mov %eax,%ds

enable_long_mode:
	MOVL	CR4, AX
	DATA32
	MOVL	$(1<<7 | 1<<5), AX		// set CR4.(PGE|PAE)
	MOVL	AX, CR4

	DATA32
	MOVL	$MSR_EFER, CX
	RDMSR
	DATA32
	ORL	$(1<<8), AX			// set MSR_EFER.LME
	WRMSR

	MOVL	CR0, AX
	DATA32
	ORL	$(1<<31 | 1<<1 |1<<0), AX	// set CR0.(PG|MP|PE)
	MOVL	AX, CR0

	// set Global Descriptor Table
	DATA32
	MOVL	$(const_gdtrBaseAddress), AX

	// convert linear address to CS offset
	DATA32
	SUBL	$(const_apinitAddress), AX

	DATA32; ADDR32
	BYTE	$0x2e				// CS segment override prefix
	LGDT	(AX)

	// segment selector for GDT entry 2
	DATA32
	MOVL	$0x10, AX

	DATA32; BYTE	$0x8e; BYTE	$0xc0	// mov %eax,%es
	DATA32; BYTE	$0x8e; BYTE	$0xd0	// mov %eax,%ss
	DATA32; BYTE	$0x8e; BYTE	$0xd8	// mov %eax,%ds
	DATA32; BYTE	$0x8e; BYTE	$0xe0	// mov %eax,%fs
	DATA32; BYTE	$0x8e; BYTE	$0xe8	// mov %eax,%gs

	// ljmp 0x08:0x6000 (FIXME: can we return here?)
	BYTE	$0xea
	BYTE	$0x00
	BYTE	$0x60
	BYTE	$0x08
	BYTE	$0x00

	// force alignment padding
	BYTE	$0xcc
	BYTE	$0xcc
	BYTE	$0xcc
	BYTE	$0xcc
	BYTE	$0xcc
	BYTE	$0xcc
	BYTE	$0xcc
	BYTE	$0xcc

// func apinit_reloc(ptr uintptr)
TEXT ·apinit_reloc(SB),$0-8
	MOVQ	$·apinit<>(SB), SI
	MOVL	ptr+0(FP), DI

	// end of function marker (alignment padding)
	MOVQ	$0xcccccccccccccccc, BX
copy_8:
	MOVQ	(SI), AX
	ADDQ	$8, SI

	CMPQ	AX, BX
	JAE	done

	MOVQ	AX, (DI)
	ADDQ	$8, DI

	JMP	copy_8
done:
	RET
