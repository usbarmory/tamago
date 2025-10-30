// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "arm64.h"

// func cache_disable()
TEXT ·cache_disable(SB),$0
	WORD	$0xd53e1000	// mrs x0, sctlr_el3
	BIC	$1<<12, R0	// disable I-cache
	BIC	$1<<2, R0	// disable D-cache
	WORD	$0xd51e1000	// msr sctlr_el3, x0
	ISB	SY
	RET

// func cache_enable()
TEXT ·cache_enable(SB),$0
	WORD	$0xd53e1000	// mrs x0, sctlr_el3
	ORR	$1<<12, R0	// enable I-cache
	ORR	$1<<2, R0	// enable D-cache
	WORD	$0xd51e1000	// msr sctlr_el3, x0
	ISB	SY
	RET
