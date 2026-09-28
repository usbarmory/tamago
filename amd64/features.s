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

TEXT xsave_enable(SB),NOSPLIT|NOFRAME,$0
	MOVL	$1, AX			// Processor Info and Feature Bits
	MOVL	$0, CX
	CPUID
	BTL	$26, CX			// check ECX.XSAVE
	JCC	done

	MOVL	CR4, BX
	ORL	$(1<<18), BX		// set CR4.OSXSAVE
	MOVL	BX, CR4

	MOVL	$0xd, AX		// XSAVE features
	MOVL	$0, CX
	CPUID
	ANDL	$0xe7, AX		// x87|SSE|AVX|opmask|ZMM_Hi256|Hi16_ZMM
	MOVL	$0, DX
	MOVL	$0, CX			// XCR0
	XSETBV				// Set Extended Control Register
done:
	RET
