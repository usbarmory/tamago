// ARM processor support
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func vfp_enable()
TEXT ·vfp_enable(SB),$0
	MRC	15, 0, R1, C1, C0, 2
	ORR	$0xf<<20, R1, R1	// Enable access for CP10 and CP11

	MCR	15, 0, R1, C1, C0, 2
	MOVW	$0, R1
	MCR	15, 0, R1, C7, C5, 4

	MOVW	$0x40000000, R3
	WORD	$0xeee83a10		// VMSR FPEXC, R3

	RET
