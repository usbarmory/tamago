// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func cache_disable()
TEXT ·cache_disable(SB),$0
	MRS	SCTLR_EL1, R0
	BIC	$1<<12, R0, R0	// Disable I-cache
	BIC	$1<<2, R0, R0	// Disable D-cache
	MSR	R0, SCTLR_EL1
	ISB	$1
	RET

// func cache_enable()
TEXT ·cache_enable(SB),$0
	MRS	SCTLR_EL1, R0
	ORR	$1<<12, R0, R0	// Enable I-cache
	ORR	$1<<2, R0, R0	// Enable D-cache
	MSR	R0, SCTLR_EL1
	ISB	$1
	RET
