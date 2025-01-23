// x86-64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func rdrand() uint32
TEXT ·rdrand(SB),$0-4
	// rdrand eax
	BYTE	$0x0f
	BYTE	$0xc7
	BYTE	$0xf0
	MOVL	AX, ret+0(FP)
	RET
