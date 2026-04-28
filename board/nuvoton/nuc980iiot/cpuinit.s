// NuMaker-IIoT-NUC980G2 board initialization
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build linkcpuinit

#include "textflag.h"

TEXT cpuinit(SB),NOSPLIT|NOFRAME,$0
	// Install minimal exception vectors (B . loops) so that any
	// abort during early init halts at a known address instead of
	// cascading through uninitialised DDR.
	MOVW	$0xEAFFFFFE, R0
	MOVW	$0, R1
	MOVW	$8, R2
vecloop:
	MOVW	R0, (R1)
	ADD	$4, R1
	SUB	$1, R2
	CMP	$0, R2
	B.NE	vecloop

	// NUC980 SoC clock gates (AIC, timers, UART0)
	BL	github·com∕usbarmory∕tamago∕soc∕nuvoton∕nuc980·EarlyInit(SB)

	// Board pin mux: GPF11 = UART0_RXD, GPF12 = UART0_TXD.
	// SYS_GPF_MFPH (0xB000009C): [15:12]=1, [19:16]=1.
	MOVW	$0xB000009C, R0
	MOVW	$0x00011000, R1
	MOVW	R1, (R0)

	// Zero BSS before reading Go variables (RamStart is in BSS and
	// may contain garbage if the DDR blob does not clear memory).
	BL	clearBSS(SB)

	// set stack pointer
	MOVW	runtime∕goos·RamStart(SB), R13
	MOVW	runtime∕goos·RamSize(SB), R1
	MOVW	runtime∕goos·RamStackOffset(SB), R2
	ADD	R1, R13
	SUB	R2, R13
	MOVW	R13, R3

	// detect HYP mode and switch to SVC if necessary
	WORD	$0xe10f0000	// mrs r0, CPSR
	AND	$0x1f, R0, R0

	CMP	$0x10, R0	// USR mode
	BL.EQ	_rt0_tamago_start(SB)

	CMP	$0x1a, R0	// HYP mode
	B.NE	after_eret

	BIC	$0x1f, R0
	ORR	$0x1d3, R0	// AIF masked, SVC mode
	MOVW	$12(R15), R14	// add lr, pc, #12 (after_eret)
	WORD	$0xe16ff000	// msr SPSR_fsxc, r0
	WORD	$0xe12ef30e	// msr ELR_hyp, lr
	WORD	$0xe160006e	// eret

after_eret:
	// enter System Mode
	WORD	$0xe321f0df	// msr CPSR_c, 0xdf

	MOVW	R3, R13
	B	_rt0_tamago_start(SB)

// clearBSS zeroes the .bss and .noptrbss sections.
//
// Note the $ prefix: MOVW $sym(SB) loads the ADDRESS of the symbol,
// while MOVW sym(SB) would load the CONTENTS at that address.
TEXT clearBSS(SB),NOSPLIT|NOFRAME,$0
	MOVW	$runtime·bss(SB), R0
	MOVW	$runtime·enoptrbss(SB), R1
	MOVW	$0, R2
bss_loop:
	CMP	R0, R1
	B.EQ	bss_done
	MOVW	R2, (R0)
	ADD	$4, R0
	B	bss_loop
bss_done:
	RET
