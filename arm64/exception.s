// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

TEXT ·handleException(SB),NOSPLIT|NOFRAME,$0
	MRS	ELR_EL1, R0
	MOVD	R0, 8(RSP)	// arg
	JMP	·systemException(SB)

// func set_vbar(addr uint64)
TEXT ·set_vbar(SB),NOSPLIT,$0
	MOVD	addr+0(FP), R0
	MSR	R0, VBAR_EL1
	RET

// func read_el() uint64
TEXT ·read_el(SB),$0-8
	MRS	CurrentEL, R0
	MOVD	R0, ret+0(FP)
	RET
