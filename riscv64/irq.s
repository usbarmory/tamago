// RISC-V 64-bit processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

// func irq_enable()
TEXT ·irq_enable(SB),NOSPLIT|NOFRAME,$0
	// enable machine level software interrupts
	MOV	$(1<<3), T0		// set MIE.MSIE
	CSRRS	T0, MIE, ZERO

	// enable global interrupts
	MOV	$(1<<3), T0		// set MSTATUS.MIE
	CSRRS	T0, MSTATUS, ZERO

	RET

// func irq_disable()
TEXT ·irq_disable(SB),NOSPLIT|NOFRAME,$0
	// disable machine level software interrupts
	MOV	$(1<<3), T0
	CSRRC	T0, MIE, ZERO

	// disable global interrupts
	MOV	$(1<<3), T0
	CSRRC	T0, MSTATUS, ZERO

	RET

// func wfi()
TEXT ·wfi(SB),NOSPLIT|NOFRAME,$0
	// wait until an interrupt is received in low-power state
	WORD $0x10500073 // wfi
	RET
