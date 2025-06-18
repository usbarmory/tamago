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

	// Set Protection Enable
	MOVL	CR0, AX
	ORL	$1, AX		// set CR0.PE
	MOVL	AX, CR0

	// Set Global Descriptor Table
	MOVL	$(const_gdtrBaseAddress), AX
	LGDT	(AX)

	MOVL	$(const_apinitAddress), AX

	PUSHW	$0x08
	PUSHW	AX

	// WiP
	// kvm_exit: vcpu 1 reason EPT_VIOLATION rip 0x1018 info1 0x0000000000000781
	// kvm_page_fault: vcpu 1 rip 0x1018 address 0x0000000000542ab8 error_code 0x781
	//PUSHW	$0x0008
	//PUSHW	$0x3aba

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
