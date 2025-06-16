// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func Move(dst uint32, src uint32)
TEXT ·Move(SB),$0-8
	MOVL	dst+0(FP), AX
	MOVL	src+4(FP), BX

	// copy src to dst
	MOVL	(AX), CX
	MOVL	CX, (AX)

	// zero out src
	MOVL	$0, CX
	MOVL	CX, (BX)

	RET

// func Write(addr uint32, val uint32)
TEXT ·Write(SB),$0-8
	MOVL	addr+0(FP), AX
	MOVL	val+4(FP), BX

	MOVL	BX, (AX)

	RET

// func Write64(addr uint64, val uint64)
TEXT ·Write64(SB),$0-16
	MOVQ	addr+0(FP), AX
	MOVQ	val+8(FP), BX

	MOVQ	BX, (AX)

	RET
