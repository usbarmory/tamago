// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

// func irq_enable()
TEXT ·irq_enable(SB),$0
	MSR	$0b1111, DAIFClr
	RET

// func irq_disable()
TEXT ·irq_disable(SB),$0
	MSR	$0b1111, DAIFSet
	RET

// func wfi()
TEXT ·wfi(SB),$0
	// wait until an interrupt is received in low-power state
	WFI
	RET

TEXT ·handleInterrupt(SB),NOSPLIT|NOFRAME,$0
	// save caller registers
	STP	(R0, R1), -(1*16)(RSP)
	STP	(R2, R3), -(2*16)(RSP)
	STP	(R4, R5), -(3*16)(RSP)
	STP	(R6, R7), -(4*16)(RSP)
	STP	(R8, R9), -(5*16)(RSP)
	STP	(R10, R11), -(6*16)(RSP)
	STP	(R12, R13), -(7*16)(RSP)
	STP	(R14, R15), -(8*16)(RSP)
	STP	(R16, R17), -(9*16)(RSP)
	STP	(R19, R20), -(10*16)(RSP)
	STP	(R21, R22), -(11*16)(RSP)
	STP	(R23, R24), -(12*16)(RSP)
	STP	(R25, R26), -(13*16)(RSP)
	STP	(R27, g), -(14*16)(RSP)
	STP	(R29, R30), -(15*16)(RSP)
	MOVD	NZCV, R0
	MOVD	R0, -(16*16)(RSP)

	// wake up IRQ handling goroutine
	MOVD	·irqHandlerG(SB), R0
	CMP	$0, R0
	BEQ	done
	CALL	runtime·WakeG(SB)

	// the IRQ handling goroutine is expected to unmask IRQs
	WORD	$0xd53e4000	// mrs x0, SPSR_EL3
	ORR	$1<<6, R0	// mask FIQs
	WORD	$0xd51e4000	// msr SPSR_EL3, x0
done:
	// restore caller registers
	MOVD	-(16*16)(RSP), R0
	MOVD	R0, NZCV
	LDP	-(15*16)(RSP), (R29, R30)
	LDP	-(14*16)(RSP), (R27, g)
	LDP	-(13*16)(RSP), (R25, R26)
	LDP	-(12*16)(RSP), (R23, R24)
	LDP	-(11*16)(RSP), (R21, R22)
	LDP	-(10*16)(RSP), (R19, R20)
	LDP	-(9*16)(RSP), (R16, R17)
	LDP	-(8*16)(RSP), (R14, R15)
	LDP	-(7*16)(RSP), (R12, R13)
	LDP	-(6*16)(RSP), (R10, R11)
	LDP	-(5*16)(RSP), (R8, R9)
	LDP	-(4*16)(RSP), (R6, R7)
	LDP	-(3*16)(RSP), (R4, R5)
	LDP	-(2*16)(RSP), (R2, R3)
	LDP	-(1*16)(RSP), (R0, R1)

	// exception return
	ERET
