// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func cache_disable()
TEXT ·cache_disable(SB),$0
	MRS	R0, SCTLR_EL1
	BIC	$1<<12, R0, R0	// Disable I-cache
	BIC	$1<<2, R0, R0	// Disable D-cache
	MSR	SCTLR_EL1, R0
	ISB
	RET

// func cache_enable()
TEXT ·cache_enable(SB),$0
	MRS	X0, SCTLR_EL1
	ORR	$1<<12, X0, X0	// Enable I-cache
	ORR	$1<<2, X0, X0	// Enable D-cache
	MSR	SCTLR_EL1, X0
	ISB
	RET
