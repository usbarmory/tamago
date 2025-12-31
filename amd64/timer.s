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
	SHLQ	$32, DX
	ORQ	DX, AX
	MOVQ	AX, ret+0(FP)
	RET

// func write_tsc_deadline(cnt uint64)
TEXT ·write_tsc_deadline(SB),$0-8
	// Intel® 64 and IA-32 Architectures Software Developer’s Manual
	// Volume 3A - 10.5.4.1 TSC-Deadline Mode
	MOVQ	$MSR_TSC_DEADLINE, CX
	MOVQ	cnt+0(FP), AX
	MOVQ	AX, DX
	SHRQ	$32, DX
	WRMSR
	RET
