// ARM processor support — ARMv5 IRQ/FIQ helpers
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// IRQ/FIQ enable/disable and WFI for ARMv5 cores that lack
// CPSIE/CPSID (ARMv6+) and the WFI opcode (ARMv6K+).

//go:build !arm.6

#include "go_asm.h"
#include "textflag.h"

// func irq_enable(spsr bool)
TEXT ·irq_enable(SB),$0
	CMP	$1, R0
	B.EQ	spsr

	WORD	$0xe10f0000 // mrs r0, CPSR
	BIC	$1<<7, R0   // unmask IRQs
	WORD	$0xe121f000 // msr CPSR_c, r0
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

	WORD	$0xe10f0000 // mrs r0, CPSR
	ORR	$1<<7, R0   // mask IRQs
	WORD	$0xe121f000 // msr CPSR_c, r0
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

	WORD	$0xe10f0000 // mrs r0, CPSR
	BIC	$1<<6, R0   // unmask FIQs
	WORD	$0xe121f000 // msr CPSR_c, r0
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

	WORD	$0xe10f0000 // mrs r0, CPSR
	ORR	$1<<6, R0   // mask FIQs
	WORD	$0xe121f000 // msr CPSR_c, r0
	RET
spsr:
	WORD	$0xe14f0000 // mrs r0, SPSR
	ORR	$1<<6, R0   // mask FIQs
	WORD	$0xe169f000 // msr SPSR, r0
	RET

// func wfi()
TEXT ·wfi(SB),$0
	// CP15 Wait For Interrupt: stalls until interrupt or FIQ.
	WORD	$0xEE070F90 // MCR p15, 0, R0, c7, c0, 4
	RET

TEXT ·irqHandler(SB),NOSPLIT|NOFRAME,$0
	// remove exception specific LR offset
	SUB	$4, R14, R14

	// save caller registers
	MOVM.DB.W	[R0-R12, R14], (R13)	// push {r0-r12, r14}

	SUB	$8, R13, R13
	MOVW	$(const_IRQ_SIGNAL), R0
	MOVW	R0, 4(R13)
	CALL	os∕signal·Relay(SB)
	ADD	$8, R13, R13

	// Mask IRQs in SPSR so the exception return restores CPSR with
	// I-bit set, preventing re-entry before ServiceInterrupts runs.
	WORD	$0xe14f0000			// mrs r0, SPSR
	ORR	$1<<7, R0			// set I bit (mask IRQs)
	WORD	$0xe169f000			// msr SPSR_fc, r0
done:
	// restore caller registers
	MOVM.IA.W	(R13), [R0-R12, R14]	// pop {r0-r12, r14}

	// restore PC from LR and mode
	MOVW.S	R14, R15
