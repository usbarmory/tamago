// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func Move(dst uint32, src uint32)
TEXT ·Move(SB),$0-8
	MOVW	dst+0(FP), R0
	MOVW	src+4(FP), R1

	// copy src to dst
	MOVW	(R1), R3
	MOVW	R3, (R0)

	// zero out src
	MOVW	$0, R3
	MOVW	R3, (R1)

	RET

// func Write(addr uint32, val uint32)
TEXT ·Write(SB),$0-8
	MOVW	addr+0(FP), R0
	MOVW	val+4(FP), R1

	MOVW	R1, (R0)

	RET

// func Write64(addr uint64, val uint64)
TEXT ·Write64(SB),$0-16
	MOVD	addr+0(FP), R0
	MOVD	val+8(FP), R1

	MOVD	R1, (R0)

	RET
