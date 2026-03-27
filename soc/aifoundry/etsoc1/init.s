// AI Foundry Minion initialization
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build linkcpuinit

#include "go_asm.h"
#include "textflag.h"

TEXT cpuinit(SB),NOSPLIT|NOFRAME,$0
	// park additional harts
	CSRRS	ZERO, MHARTID, T0
	MOV	$0, T1
	BGT	T0, T1, wait

	JMP	github·com∕usbarmory∕tamago∕riscv64·Init(SB)
wait:
	// enable machine level software interrupts
	MOV	$(1<<3), T0	// set MIE.MSIE
	CSRRS	T0, MIE, ZERO
wfi:
	// wait until an interrupt is received in low-power state
	WORD	$0x10500073	// wfi

	// get hartid bit position
	CSRRS	ZERO, MHARTID, T1
	MOV	$1, T2
	SLL	T1, T2, T2
	OR	T2, T2

	// clear IPI
	MOV	$(const_IPI_TRIGGER_CLEAR), T0
	MOV	T2, (T0)

	JMP	wfi
