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

	// Disable interrupts
	CLI

	MOVL	CR0, AX
	ORL	$1, AX		// set CR0.PE
	MOVL	AX, CR0

	// TODO: lgdt

	MOVL	$(const_trampoline), AX
	PUSHW	$0x8
	PUSHW	AX
	RETFL

TEXT ·setup_ap_trampoline(SB),NOSPLIT|NOFRAME,$0
	// FIXME: NOP
	MOVB	$0xf4, AX		// HLT
	MOVB	AX, (DI)
	ADDQ	$1, DI

	// FIXME: is this a reliable marker?
	MOVQ	$0xcccccccccccccccc, BX

	MOVQ	$·apinit<>(SB), SI
	MOVQ	$(const_trampoline), DI
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
