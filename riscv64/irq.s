// RISC-V 64-bit processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"
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
	WORD	$0x10500073 // wfi
	RET

TEXT ·handleInterrupt(SB),NOSPLIT|NOFRAME,$0
	// save caller registers
	MOV	X1, -2*8(SP)
	MOV	X3, -3*8(SP)
	MOV	TP, -4*8(SP)
	MOV	X5, -5*8(SP)
	MOV	X6, -6*8(SP)
	MOV	X7, -7*8(SP)
	MOV	X8, -8*8(SP)
	MOV	X9, -9*8(SP)
	MOV	X10, -10*8(SP)
	MOV	X11, -11*8(SP)
	MOV	X12, -12*8(SP)
	MOV	X13, -13*8(SP)
	MOV	X14, -14*8(SP)
	MOV	X15, -15*8(SP)
	MOV	X16, -16*8(SP)
	MOV	X17, -17*8(SP)
	MOV	X18, -18*8(SP)
	MOV	X19, -19*8(SP)
	MOV	X20, -20*8(SP)
	MOV	X21, -21*8(SP)
	MOV	X22, -22*8(SP)
	MOV	X23, -23*8(SP)
	MOV	X24, -24*8(SP)
	MOV	X25, -25*8(SP)
	MOV	X26, -26*8(SP)
	MOV	g,   -27*8(SP)
	MOV	X28, -28*8(SP)
	MOV	X29, -29*8(SP)
	MOV	X30, -30*8(SP)
	MOV	X31, -31*8(SP)

	SUB	$(32*8), SP
	MOV	$(const_IRQ_SIGNAL), T0
	MOV	T0, 8(SP)
	CALL	os∕signal·Relay(SB)
	ADD	$(32*8), SP

	// the IRQ handling goroutine is expected to unmask IRQs
	CALL	·irq_disable(SB)
done:
	// restore caller registers
	MOV	-2*8(SP), X1
	MOV	-3*8(SP), X3
	MOV	-4*8(SP), TP
	MOV	-5*8(SP), X5
	MOV	-6*8(SP), X6
	MOV	-7*8(SP), X7
	MOV	-8*8(SP), X8
	MOV	-9*8(SP), X9
	MOV	-10*8(SP), X10
	MOV	-11*8(SP), X11
	MOV	-12*8(SP), X12
	MOV	-13*8(SP), X13
	MOV	-14*8(SP), X14
	MOV	-15*8(SP), X15
	MOV	-16*8(SP), X16
	MOV	-17*8(SP), X17
	MOV	-18*8(SP), X18
	MOV	-19*8(SP), X19
	MOV	-20*8(SP), X20
	MOV	-21*8(SP), X21
	MOV	-22*8(SP), X22
	MOV	-23*8(SP), X23
	MOV	-24*8(SP), X24
	MOV	-25*8(SP), X25
	MOV	-26*8(SP), X26
	MOV	-27*8(SP), g
	MOV	-28*8(SP), X28
	MOV	-29*8(SP), X29
	MOV	-30*8(SP), X30
	MOV	-31*8(SP), X31

	// exception return
	RET
