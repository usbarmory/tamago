// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func Move(dst uint32, src uint32)
TEXT ·Move(SB),$0-8
	MOV	dst+0(FP), T0
	MOV	src+4(FP), T1

	// copy src to dst
	MOV	(T1), T3
	MOV	T3, (T0)

	// zero out src
	MOV	$0, T3
	MOV	T3, (T1)

	RET
