// RISC-V 64-bit processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

TEXT ·Init(SB),NOSPLIT|NOFRAME,$0
	// disable interrupts
	MOV	$0, T0
	CSRRW	T0, SIE, ZERO
	CSRRW	T0, MIE, ZERO
	MOV	$0x7fff, T0
	CSRRC	T0, MSTATUS, ZERO

	// enable FPU
	MOV	$(1<<13), T0
	CSRRS	T0, MSTATUS, ZERO

	// set stack pointer
	MOV	runtime∕goos·RamStart(SB), X2
	MOV	runtime∕goos·RamSize(SB), T1
	MOV	runtime∕goos·RamStackOffset(SB), T2
	ADD	T1, X2
	SUB	T2, X2

	JMP	_rt0_tamago_start(SB)

// func exit(int32)
TEXT ·exit(SB),$0-8
	// wait forever in low-power state
	WORD	$0x10500073 // wfi
