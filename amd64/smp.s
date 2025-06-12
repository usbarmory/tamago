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
	CLI

TEXT ·setup_ap_trampoline(SB),NOSPLIT|NOFRAME,$0
	MOVQ	$(const_trampoline), DI

	MOVB	$0xf4, AX		// HLT
	MOVB	AX, (DI)
	ADDQ	$1, DI

	// TODO: Use Go assembly
	// TODO: lgdt

	// set CR0.PE
	MOVQ	$0xc0220f010cc0200f, AX	// mov %cr0,%rax
	MOVQ	AX, (DI)		// or  $0x1,%al
	ADDQ	$8, DI			// mov %rax,%cr0

	MOVQ	$0x66086a00008000b8, AX	// mov 0x8000,%eax
	MOVQ	AX, (DI)		// push $0x8
	ADDQ	$8, DI			// push %ax
	MOVQ	$0xcb50, AX		// lret
	MOVQ	AX, (DI)

	//MOVL	$·apinit<>(SB), AX
	//MOVL	AX, (DI)
	//ADDL	$4, DI

	RET
