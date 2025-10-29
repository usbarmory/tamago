// AMD64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

// func cpuid(eaxArg, ecxArg uint32) (eax, ebx, ecx, edx uint32)
TEXT ·cpuid(SB),NOSPLIT,$0-24
	MOVL eaxArg+0(FP), AX
	MOVL ecxArg+4(FP), CX
	CPUID
	MOVL AX, eax+8(FP)
	MOVL BX, ebx+12(FP)
	MOVL CX, ecx+16(FP)
	MOVL DX, edx+20(FP)
	RET

TEXT sse_enable(SB),NOSPLIT|NOFRAME,$0
	MOVL	CR0, AX
	MOVL	CR4, BX

	ANDL	$~(1<<2), AX		// clear CR0.EM
	ORL	$(1<<1), AX		//   set CR0.MP
	ORL	$(1<<10 | 1<<9), BX	//   set CR4.(OSXMMEXCPT|OSFXSR)

	MOVL	AX, CR0
	MOVL	BX, CR4

	RET
