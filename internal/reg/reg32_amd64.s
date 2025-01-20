// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func Move(dst uint32, src uint32)
TEXT ·Move(SB),$0-8
	MOVW	dst+0(FP), AX
	MOVW	src+4(FP), BX

	// copy src to dst
	MOVW	(AX), CX
	MOVW	CX, (AX)

	// zero out src
	MOVW	$0, CX
	MOVW	CX, (BX)

	RET
