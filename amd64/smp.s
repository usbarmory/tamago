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
	// 16-bit real mode: operand/address size override prefixes required

	// WiP
	//HLT

	// disable interrupts
	CLI

	// set Protection Enable
	MOVL	CR0, AX
	ORL	$1, AX				// set CR0.PE
	MOVL	AX, CR0

	// set Global Descriptor Table
	BYTE	$0x66				// 32-bit operand size override prefix
	BYTE	$0x67				// 32-bit address size override prefix
	MOVL	$(const_gdtrBaseAddress), AX
	BYTE	$0x66				// 32-bit operand size override prefix
	BYTE	$0x67				// 32-bit address size override prefix
	LGDT	(AX)

	// set far return target
	BYTE	$0x66				// 32-bit operand size override prefix
	MOVL	$(const_apinitAddress), AX

	// jump to target in protected mode
	PUSHW	$0x08
	BYTE	$0x66				// 32-bit operand size override prefix
	PUSHQ	AX
	RETFL

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
