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
	// 16-bit real mode

	// WiP
	//HLT

	// Disable interrupts
	CLI

	MOVL	CR0, AX
	ORL	$1, AX		// set CR0.PE
	MOVL	AX, CR0

	MOVL	$(const_gdtBaseAddress+24), AX
	LGDT	(AX)

	MOVQ	$(const_apinitAddress), AX
	PUSHQ	$0x08
	PUSHQ	AX
	RETFW	// FIXME: lret vs lretw vs lretq

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
