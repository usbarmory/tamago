// ARM processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"

// func vfp_enable()
TEXT ·vfp_enable(SB),$0
	MRC	15, 0, R1, C1, C0, 2
	ORR	$(0xf << 20), R1, R1	// enable CP10 and CP11 access
	MCR	15, 0, R1, C1, C0, 2

	MOVW	$0, R1
	MCR	15, 0, R1, C7, C5, 4	// CP15ISB

	MOVW	$(1<<const_FPEXC_EN), R3
	WORD	$0xeee83a10		// vmsr fpexc, r3

	RET
