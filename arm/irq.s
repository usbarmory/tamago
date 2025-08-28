// ARM processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

// func irq_enable(spsr bool)
TEXT ·irq_enable(SB),$0
	CMP	$1, R0
	B.EQ	spsr

	WORD	$0xf1080080 // cpsie i
	RET
spsr:
	WORD	$0xe14f0000 // mrs r0, SPSR
	BIC	$1<<7, R0   // unmask IRQs
	WORD	$0xe169f000 // msr SPSR, r0
	RET

// func irq_disable(spsr bool)
TEXT ·irq_disable(SB),$0
	CMP	$1, R0
	B.EQ	spsr

	WORD	$0xf10c0080 // cpsid i
	RET
spsr:
	WORD	$0xe14f0000 // mrs r0, SPSR
	ORR	$1<<7, R0   // mask IRQs
	WORD	$0xe169f000 // msr SPSR, r0
	RET

// func fiq_enable(spsr bool)
TEXT ·fiq_enable(SB),$0
	CMP	$1, R0
	B.EQ	spsr

	WORD	$0xf1080040 // cpsie f
	RET
spsr:
	WORD	$0xe14f0000 // mrs r0, SPSR
	BIC	$1<<6, R0   // unmask FIQs
	WORD	$0xe169f000 // msr SPSR, r0
	RET

// func fiq_disable(spsr bool)
TEXT ·fiq_disable(SB),$0
	CMP	$1, R0
	B.EQ	spsr

	WORD	$0xf10c0040 // cpsid f
	RET
spsr:
	WORD	$0xe14f0000 // mrs r0, SPSR
	ORR	$1<<6, R0   // mask FIQs
	WORD	$0xe169f000 // msr SPSR, r0
	RET

TEXT ·irqHandler(SB),NOSPLIT|NOFRAME,$0
	// remove exception specific LR offset
	SUB	$4, R14, R14

	// save caller registers
	MOVM.DB.W	[R0-R12, R14], (R13)	// push {r0-r12, r14}

	// wake up IRQ handling goroutine
	MOVW	·irqHandlerG(SB), R0
	CMP	$0, R0
	B.EQ	done
	CALL	runtime·WakeG(SB)

	// the IRQ handling goroutine is expected to unmask IRQs
	WORD	$0xe14f0000			// mrs r0, SPSR
	ORR	$1<<7, R0			// mask IRQs
	WORD	$0xe169f000			// msr SPSR, r0
done:
	// restore caller registers
	MOVM.IA.W	(R13), [R0-R12, R14]	// pop {r0-r12, r14}

	// restore PC from LR and mode
	MOVW.S	R14, R15
