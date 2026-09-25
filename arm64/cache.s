// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "arm64.h"

// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile
// D12.2.100 SCTLR_EL1, System Control Register (EL1)

// func cache_disable()
TEXT ·cache_disable(SB),$0
	MRS	SCTLR_EL1, R0
	BIC	$1<<12, R0	// disable I-cache
	BIC	$1<<2, R0	// disable D-cache
	MSR	R0, SCTLR_EL1
	ISB	SY
	RET

// func cache_enable()
TEXT ·cache_enable(SB),$0
	MRS	SCTLR_EL1, R0
	ORR	$1<<12, R0	// enable I-cache
	ORR	$1<<2, R0	// enable D-cache
	MSR	R0, SCTLR_EL1
	ISB	SY
	RET

// func cache_line_size() uint64
TEXT ·cache_line_size(SB),$0-8
	// CTR_EL0.DminLine is the log2 word count of the smallest data line
	MRS	CTR_EL0, R0
	UBFX	$16, R0, $4, R0
	MOVD	$4, R1
	LSL	R0, R1, R0
	MOVD	R0, ret+0(FP)
	RET

// func cache_invalidate_range(start uint64, end uint64, line uint64)
TEXT ·cache_invalidate_range(SB),$0-24
	MOVD	start+0(FP), R0
	MOVD	end+8(FP), R1
	MOVD	line+16(FP), R2
invalidate:
	DC	IVAC, R0
	ADD	R2, R0, R0
	CMP	R1, R0
	BLO	invalidate
	DSB	SY
	RET
