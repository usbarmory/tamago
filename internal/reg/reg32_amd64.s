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
