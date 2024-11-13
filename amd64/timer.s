// x86-64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func read_tsc() int64
TEXT ·read_tsc(SB),$0-8
	// rdtsc
	BYTE	$0x0f
	BYTE	$0x31
	MOVL	AX, ret+0(FP)
	MOVL	DX, ret+4(FP)
	RET
