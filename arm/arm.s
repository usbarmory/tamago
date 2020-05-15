// ARM processor support
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func read_cpsr() uint32
TEXT ·read_cpsr(SB),$0
	WORD	$0xe10f0000		// mrs r0, CPSR
	MOVW	R0, ret+0(FP)

	RET

// func read_scr() uint32
TEXT ·read_scr(SB),$0
	MRC	15, 0, R0, C1, C1, 0	// Read SCR into R0
	MOVW	R0, ret+0(FP)

	RET
