// First-fit memory allocator for DMA buffers
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func cache_line_size() uint64
TEXT ·cache_line_size(SB),$0-8
	// CTR_EL0.DminLine is the log2 word count of the smallest data line
	MRS	CTR_EL0, R0
	UBFX	$16, R0, $4, R0
	MOVD	$4, R1
	LSL	R0, R1, R0
	MOVD	R0, ret+0(FP)
	RET

// func cache_clean(start uint64, end uint64, line uint64)
TEXT ·cache_clean(SB),$0-24
	MOVD	start+0(FP), R0
	MOVD	end+8(FP), R1
	MOVD	line+16(FP), R2
clean:
	DC	CVAC, R0
	ADD	R2, R0, R0
	CMP	R1, R0
	BLO	clean
	DSB	$0b1111	// SY
	RET

// func cache_invalidate(start uint64, end uint64, line uint64)
TEXT ·cache_invalidate(SB),$0-24
	MOVD	start+0(FP), R0
	MOVD	end+8(FP), R1
	MOVD	line+16(FP), R2
invalidate:
	DC	IVAC, R0
	ADD	R2, R0, R0
	CMP	R1, R0
	BLO	invalidate
	DSB	$0b1111	// SY
	RET
