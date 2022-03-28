// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
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
