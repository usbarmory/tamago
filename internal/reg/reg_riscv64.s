// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func Move(dst uint32, src uint32)
TEXT ·Move(SB),$0-8
	MOVW	dst+0(FP), T0
	MOVW	src+4(FP), T1

	// copy src to dst
	MOVW	(T1), T3
	MOVW	T3, (T0)

	// zero out src
	MOV	$0, T3
	MOVW	T3, (T1)

	RET

// func Write(addr uint32, val uint32)
TEXT ·Write(SB),$0-8
	MOVW	addr+0(FP), T0
	MOVW	val+4(FP), T1

	MOVW	T1, (T0)

	RET

// func Write64(addr uint64, val uint64)
TEXT ·Write64(SB),$0-16
	MOV	addr+0(FP), T0
	MOV	val+8(FP), T1

	MOV	T1, (T0)

	RET
