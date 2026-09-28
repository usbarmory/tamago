// AMD64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "amd64.h"
#include "textflag.h"

// func read_cr0() uint64
TEXT ·read_cr0(SB),$0-8
	MOVQ	CR0, AX
	MOVQ	AX, ret+0(FP)
	RET

// func write_cr0(val uint64)
TEXT ·write_cr0(SB),$0-8
	MOVQ	val+0(FP), AX
	MOVQ	AX, CR0

	RET

// func read_cr3() uint64
TEXT ·read_cr3(SB),$0-8
	MOVQ	CR3, AX
	MOVQ	AX, ret+0(FP)
	RET

// func set_ext_pdpt(index int, flags uint64)
TEXT ·set_ext_pdpt(SB),NOSPLIT,$0-16
	// detect cpuinit override
	MOVQ	CR3, AX
	SHRQ	$12, AX
	CMPQ	AX, $(PML4T>>12)
	JNE	done

	MOVQ	index+0(FP), AX
	MOVQ	flags+8(FP), BX

	// PDPT[index]: index << 30 | flags

	MOVQ	AX, CX
	SHLQ	$30, CX
	ORQ	BX, CX

	MOVL	$PDPT, DI
	LEAQ	(DI)(AX*8), DI
	MOVQ	CX, (DI)
done:
	RET

// func flush_tlb()
TEXT ·flush_tlb(SB),NOSPLIT,$0
	MOVQ	CR3, AX
	MOVQ	AX, CR3
	RET
