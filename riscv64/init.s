// RISC-V 64-bit processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkcpuinit

#include "textflag.h"

TEXT cpuinit(SB),NOSPLIT|NOFRAME,$0
	// disable interrupts
	MOV	$0, T0
	CSRRW	ZERO, SIE, T0
	CSRRW	ZERO, MIE, T0
	MOV	$0x7FFF, T0
	CSRRC	ZERO, MSTATUS, T0

	// enable FPU
	MOV	$(1<<13), T0
	CSRRS	ZERO, MSTATUS, T0

	// set stack pointer
	MOV	runtime∕goos·RamStart(SB), X2
	MOV	runtime∕goos·RamSize(SB), T1
	MOV	runtime∕goos·RamStackOffset(SB), T2
	ADD	T1, X2
	SUB	T2, X2

	JMP	_rt0_tamago_start(SB)
