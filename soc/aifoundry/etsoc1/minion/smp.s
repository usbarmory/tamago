// AI Foundry Minion initialization
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

TEXT ·apstart(SB),NOSPLIT|NOFRAME,$0
	// enable FPU
	MOV	$(1<<13), T0
	CSRRS	T0, MSTATUS, ZERO

	// enable machine level software interrupts
	MOV	$(1<<3), T0	// set MIE.MSIE
	CSRRS	T0, MIE, ZERO
wfi:
	// wait IPI from [schedule]
	WORD	$0x10500073	// wfi

	// compute taskAddress
	CSRRS	ZERO, MHARTID, T0
	MOV	$(task__size), T1
	MUL	T0, T1, T0
	MOV	·taskBase(SB), T1
	ADD	T0, T1, T0

	MOV	task_sp(T0), SP
	MOV	task_gp(T0), g
	MOV	task_pc(T0), T1
	BEQ	T1, ZERO, done

	// call task target
	CALL	T1
done:
	// get hartid bit position
	CSRRS	ZERO, MHARTID, T1
	MOV	$1, T2
	SLL	T1, T2, T2
	OR	T2, T2

	// clear IPI
	MOV	$(const_IPI_TRIGGER_CLEAR), T0
	MOV	T2, (T0)

	JMP	wfi
