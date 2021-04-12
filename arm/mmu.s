// ARM processor support
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

// func set_ttbr0(addr uint32)
TEXT ·set_ttbr0(SB),NOSPLIT,$0-4
	// Data Memory Barrier
	MOVW	$0, R0
	MCR	15, 0, R0, C7, C10, 5

	// Invalidate Instruction Cache + DSB
	MOVW	$0, R1
	MCR	15, 0, R1, C7, C5, 0
	MCR	15, 0, R1, C7, C10, 4

	MOVW	addr+0(FP), R0

	// Invalidate unified TLB
	MCR	15, 0, R0, C8, C7, 0	// TLBIALL

	// Set TTBR0
	MCR	15, 0, R0, C2, C0, 0

	// Use TTBR0 for translation table walks
	MOVW	$0x0, R0
	MCR	15, 0, R0, C2, C0, 2

	// Set Domain Access
	MOVW	$0x1, R0
	MCR	15, 0, R0, C3, C0, 0

	// Invalidate Instruction Cache + DSB
	MOVW	$0, R0
	MCR	15, 0, R0, C7, C5, 0
	MCR	15, 0, R0, C7, C10, 4

	// Enable MMU
	MRC	15, 0, R0, C1, C0, 0
	ORR	$0x1, R0
	MCR	15, 0, R0, C1, C0, 0

	RET
