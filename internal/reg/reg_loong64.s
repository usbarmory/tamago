// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func Move(dst uint32, src uint32)
TEXT ·Move(SB),$0-8
	MOVWU	dst+0(FP), R4
	MOVWU	src+4(FP), R5

	// copy src to dst
	MOVWU	(R5), R6
	MOVWU	R6, (R4)

	// zero out src
	MOVWU	R0, (R5)

	RET

// func Write(addr uint32, val uint32)
TEXT ·Write(SB),$0-8
	MOVWU	addr+0(FP), R4
	MOVWU	val+4(FP), R5

	MOVWU	R5, (R4)

	RET

// func Write64(addr uint64, val uint64)
TEXT ·Write64(SB),$0-16
	MOVV	addr+0(FP), R4
	MOVV	val+8(FP), R5

	MOVV	R5, (R4)

	RET
