// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkcpuinit

#include "arm64.h"
#include "textflag.h"

TEXT cpuinit(SB),NOSPLIT|NOFRAME,$0
	// get current exception level
	MRS	CurrentEL, R0
	LSR	$2, R0, R0
	AND	$0b11, R0, R0

	// ensure we are running at EL3
	CMP	$3, R0
	BNE	exit

	// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture
	// profile.

	// D12.2.102 SCTLR_EL3, System Control Register (EL3)
	WORD	$0xd53e1000	// mrs x0, sctlr_el3
	BIC	$1<<1, R0	// clear A bit
	BIC	$1<<0, R0	// clear M bit
	WORD	$0xd51e1000	// msr sctlr_el3, x0
	ISB	SY

	// D12.2.99 SCR_EL3, Secure Configuration Register
	WORD	$0xd53e1100	// mrs x0, scr_el3
	ORR	$1<<3, R0	// set EA bit
	ORR	$1<<2, R0	// set FIQ bit
	ORR	$1<<1, R0	// set IRQ bit
	WORD	$0xd51e1100	// msr scr_el3, x0
	ISB	SY

	MOVD	runtime∕goos·RamStart(SB), R1
	MOVD	R1, RSP
	MOVD	runtime∕goos·RamSize(SB), R1
	MOVD	runtime∕goos·RamStackOffset(SB), R2
	ADD	R1, RSP
	SUB	R2, RSP

	B	_rt0_tamago_start(SB)

exit:
	JMP	·exit(SB)
