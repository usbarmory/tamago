// x86-64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#define MSR_TSC_DEADLINE 0x000006e0

// func read_tsc() uint64
TEXT ·read_tsc(SB),$0-8
	RDTSC
	MOVL	AX, ret+0(FP)
	MOVL	DX, ret+4(FP)
	RET

// func write_tsc_deadline(val uint64)
TEXT ·write_tsc_deadline(SB),$0-8
	MOVL	val+0(FP), AX
	MOVL	val+4(FP), DX
	MOVL	$MSR_TSC_DEADLINE, CX
	WRMSR
	RET
