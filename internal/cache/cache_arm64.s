// Data cache maintenance by address range
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "../../arm64/arm64.h"

// func cache_type() uint64
TEXT ·cache_type(SB),$0-8
	MRS	CTR_EL0, R0
	MOVD	R0, ret+0(FP)
	RET

// func cache_line_size() uint64
TEXT ·cache_line_size(SB),$0-8
	MRS	CTR_EL0, R0
	UBFX	$16, R0, $4, R0
	MOVD	$4, R1
	LSL	R0, R1, R0
	MOVD	R0, ret+0(FP)
	RET

// func cache_clean_data(start, end, line uint64)
TEXT ·cache_clean_data(SB),$0-24
	MOVD	start+0(FP), R0
	MOVD	end+8(FP), R1
	MOVD	line+16(FP), R2
clean:
	DC	CVAC, R0
	ADD	R2, R0, R0
	CMP	R1, R0
	BLO	clean
	DSB	SY
	RET

// func cache_clean_invalidate_data(start, end, line uint64)
TEXT ·cache_clean_invalidate_data(SB),$0-24
	MOVD	start+0(FP), R0
	MOVD	end+8(FP), R1
	MOVD	line+16(FP), R2
clean_invalidate:
	DC	CIVAC, R0
	ADD	R2, R0, R0
	CMP	R1, R0
	BLO	clean_invalidate
	DSB	SY
	RET

// func cache_invalidate_data(start, end, line uint64)
TEXT ·cache_invalidate_data(SB),$0-24
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
