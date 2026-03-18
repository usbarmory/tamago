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
	MRS	CurrentEL, R0
	LSR	$2, R0, R0
	AND	$0b11, R0, R0

	// already at EL1
	CMP	$1, R0
	BEQ	init

	// While tamago has been tested in Secure EL3, we drop to Non-secure
	// EL1 to ease chain loading from TF-A or bootloaders, as on AArch64
	// the OS is expected to run at this level.
	//
	// Future tamago unikernels for TF-A replacement or Secure monitors can
	// branch here to remain in EL3 provided that _EL1 register access is
	// replaced with _EL3.

	// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture
	// profile.

	// D12.2.99 SCR_EL3, Secure Configuration Register
	MOVD	$0, R0
	ORR	$1<<10, R0	// set lower levels as AArch64
	ORR	$1<<5, R0	// set reserved bit
	ORR	$1<<4, R0	// set reserved bit
	ORR	$1<<0, R0	// set Non-secure state
	WORD	$0xd51e1100	// msr scr_el3, x0

	// D12.2.44 HCR_EL2, Hypervisor Configuration Register
	MOVD	$1<<31, R0	// set EL1 level as AArch64
	WORD	$0xd51c1100	// msr HCR_EL2, x0

	// C5.2.19 SPSR_EL3, Saved Program Status Register (EL3)
	MOVD	$0, R0
	ORR	$0b1111<<6, R0	// mask exceptions/interrupts
	ORR	$0b0101<<0, R0	// set EL1h
	WORD	$0xd51e4000	// msr SPSR_EL3, x0

	// drop to EL1
	MOVD	$·cpuinit_el1(SB), R0
	WORD	$0xd51e4020	// msr ELR_EL3, x0
	ISB	SY
	ERET
init:
	B	·cpuinit_el1(SB)

TEXT ·cpuinit_el1(SB),NOSPLIT|NOFRAME,$0
	// D12.2.100 SCTLR_EL1, System Control Register (EL1)
	MRS	SCTLR_EL1, R0
	BIC	$1<<1, R0	// clear A bit
	BIC	$1<<0, R0	// clear M bit
	MSR	R0, SCTLR_EL1
	ISB	SY

	// set stack pointer
	MOVD	runtime∕goos·RamStart(SB), R1
	MOVD	R1, RSP
	MOVD	runtime∕goos·RamSize(SB), R1
	MOVD	runtime∕goos·RamStackOffset(SB), R2
	ADD	R1, RSP
	SUB	R2, RSP

	B	_rt0_tamago_start(SB)
