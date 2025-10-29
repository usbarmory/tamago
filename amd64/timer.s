// AMD64 processor support
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

// func write_tsc_deadline(cnt uint64)
TEXT ·write_tsc_deadline(SB),$0-8
	// Intel® 64 and IA-32 Architectures Software Developer’s Manual
	// Volume 3A - 10.5.4.1 TSC-Deadline Mode
	MOVL	cnt+0(FP), AX
	MOVL	cnt+4(FP), DX
	MOVL	$MSR_TSC_DEADLINE, CX
	WRMSR
	RET
