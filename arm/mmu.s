// ARM processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

// func flush_tlb()
TEXT ·flush_tlb(SB),NOSPLIT,$0
	MOVW	$0, R0

	// data Memory Barrier
	MCR	15, 0, R0, C7, C10, 5

	// invalidate Instruction Cache
	MCR	15, 0, R0, C7, C5, 0

	// data Synchronization Barrier
	MCR	15, 0, R0, C7, C10, 4

	// invalidate unified TLB
	MCR	15, 0, R0, C8, C7, 0

	RET

// func set_ttbr0(addr uint32)
TEXT ·set_ttbr0(SB),NOSPLIT,$0-4
	// set TTBR0
	MOVW	addr+0(FP), R0
	MCR	15, 0, R0, C2, C0, 0

	// use TTBR0 for translation table walks
	MOVW	$0, R0
	MCR	15, 0, R0, C2, C0, 2

	// set Domain Access
	MOVW	$1, R0
	MCR	15, 0, R0, C3, C0, 0

	// enable MMU
	MRC	15, 0, R0, C1, C0, 0
	ORR	$1, R0
	MCR	15, 0, R0, C1, C0, 0

	CALL	·flush_tlb(SB)

	RET
