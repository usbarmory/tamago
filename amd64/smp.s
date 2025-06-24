// x86-64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

// FIXME: WiP SMP

TEXT ·apinit<>(SB),NOSPLIT|NOFRAME,$0
	// 16-bit real mode, the following prefixes are required:
	//   0x66 32-bit operand size override prefix
	//   0x67 32-bit address size override prefix

	// disable interrupts
	CLI

	// mov %cs,%ax
	BYTE	$0x8c
	BYTE	$0xc8
	// mov %eax,%ds
	BYTE	$0x8e
	BYTE	$0xd8

	// set Protection Enable
	MOVL	CR0, AX
	ORL	$1, AX			// set CR0.PE
	MOVL	AX, CR0

	// set Global Descriptor Table

	BYTE	$0x66
	BYTE	$0x67
	MOVL	$(const_gdtrBaseAddress), AX

	// convert linear address to CS offset
	BYTE	$0x66
	BYTE	$0x67
	SUBL	$(const_apinitAddress), AX

	BYTE	$0x67
	BYTE	$0x2e			// CS segment override prefix
	LGDT	(AX)

	BYTE	$0x66
	MOVL	$0xf000, SP

	// segment selector for GDT entry 2
	BYTE	$0x66
	MOVL	$0x10, AX

	// mov %eax,%es
	BYTE	$0x66
	BYTE	$0x8e
	BYTE	$0xc0

	// mov %eax,%ss
	BYTE	$0x66
	BYTE	$0x8e
	BYTE	$0xd0

	// mov %eax,%ds
	BYTE	$0x66
	BYTE	$0x8e
	BYTE	$0xd8

	// mov %eax,%fs
	BYTE	$0x66
	BYTE	$0x8e
	BYTE	$0xe0

	// mov %eax,%gs
	BYTE	$0x66
	BYTE	$0x8e
	BYTE	$0xe8

	// jump to target in protected mode
	// ljmp 0x08:0x6000 (FIXME)
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
