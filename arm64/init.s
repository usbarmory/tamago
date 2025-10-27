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
	MRS	CurrentEL, R0
	LSR	$2, R0, R0
	AND	$0b11, R0, R0

	// WiP: run in EL3 for now with MMU disabled
	WORD	$0xd53e1000	// mrs x0, sctlr_el3
	BIC	$(1<<0), R0, R0	// clear M bit
	WORD	$0xd51e1000	// msr sctlr_el3, r0
	JMP	after_eret

	// switch to EL1 if necessary
	CMP	$0b01, R0
	BEQ	after_eret

	MOVD	$_rt0_tamago_start(SB), R0
	WORD	$0xd51e4020	// msr elr_el3, x0
	JMP	switch<>(SB)

after_eret:
	B	_rt0_tamago_start(SB)

TEXT switch<>(SB),NOSPLIT|NOFRAME,$0
	WORD	$0xd53e1100	// mrs x0, scr_el3
	ORR	$0x400, R0, R0	// set RW
	WORD	$0xd51e1100	// msr scr_el3, r0

	WORD	$0xd53e4000	// mrs x0, spsr_el3
	ORR	$0x1c5, R0, R0	// AIF masked, EL1 mode
	WORD	$0xd51e4000	// msr spsr_el3, r0
	ERET
