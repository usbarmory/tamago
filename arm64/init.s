// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkcpuinit

#include "textflag.h"

TEXT cpuinit(SB),NOSPLIT|NOFRAME,$0
	// get current exception level
	MRS     CurrentEL, R0
	LSR     $2, R0, R0
	AND     $0b11, R0, R0

	// ensure we are running at EL3
	CMP	$3, R0
	BNE	exit

	WORD	$0xd53e1000	// mrs x0, sctlr_el3
	BIC	$(1<<1), R0	// clear A bit
	BIC	$(1<<0), R0	// clear M bit
	WORD	$0xd51e1000	// msr sctlr_el3, x0
	ISB	$15

	B	_rt0_tamago_start(SB)

exit:
	JMP	·exit(SB)
