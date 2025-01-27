// x86-64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

// Interrupt Descriptor Table
GLOBL	idt<>(SB),8,$(const_vectors*16)

DATA	idtptr<>+0x00(SB)/2, $(const_vectors*16-1)	// IDT Limit
DATA	idtptr<>+0x02(SB)/8, $idt<>(SB)			// IDT Base Address
GLOBL	idtptr<>(SB),8,$(2+8)

TEXT ·irqHandler(SB),NOSPLIT|NOFRAME,$0
	// save caller registers
	MOVQ	R15, r15-(14*8+8)(SP)
	MOVQ	R14, r14-(13*8+8)(SP)
	MOVQ	R13, r13-(12*8+8)(SP)
	MOVQ	R12, r12-(11*8+8)(SP)
	MOVQ	R11, r11-(10*8+8)(SP)
	MOVQ	R10, r10-(9*8+8)(SP)
	MOVQ	R9, r9-(8*8+8)(SP)
	MOVQ	R8, r8-(7*8+8)(SP)
	MOVQ	DI, di-(6*8+8)(SP)
	MOVQ	SI, si-(5*8+8)(SP)
	MOVQ	BP, bp-(4*8+8)(SP)
	MOVQ	BX, bx-(3*8+8)(SP)
	MOVQ	DX, dx-(2*8+8)(SP)
	MOVQ	CX, cx-(1*8+8)(SP)
	MOVQ	AX, ax-(0*8+8)(SP)

	// the IRQ handling goroutine is expected to unmask IRQs
	MOVQ	rflags-(-16)(SP), AX
	ANDL	$~(1<<9), AX		// clear RFLAGS.IF
	MOVQ	AX, rflags-(-16)(SP)

	SUBQ	$(15*8+8), SP

	MOVQ	·irqHandlerG(SB), AX
	CMPQ	AX, $0
	JE	done
	CALL	runtime·WakeG(SB)
done:
	ADDQ	$(15*8+8), SP

	// restore caller registers
	MOVQ	ax-(0*8+8)(SP), AX
	MOVQ	cx-(1*8+8)(SP), CX
	MOVQ	dx-(2*8+8)(SP), DX
	MOVQ	bx-(3*8+8)(SP), BX
	MOVQ	bp-(4*8+8)(SP), BP
	MOVQ	si-(5*8+8)(SP), SI
	MOVQ	di-(6*8+8)(SP), DI
	MOVQ	r8-(7*8+8)(SP), R8
	MOVQ	r9-(8*8+8)(SP), R9
	MOVQ	r10-(9*8+8)(SP), R10
	MOVQ	r11-(10*8+8)(SP), R11
	MOVQ	r12-(11*8+8)(SP), R12
	MOVQ	r13-(12*8+8)(SP), R13
	MOVQ	r14-(13*8+8)(SP), R14
	MOVQ	r15-(14*8+8)(SP), R15

	// return to caller
	IRETQ

// func load_idt() (idt uintptr, irqHandler uintptr)
TEXT ·load_idt(SB),$0-16
	MOVQ	$idtptr<>(SB), AX
	LIDT (AX)

	MOVQ	$idt<>(SB), AX
	MOVQ	AX, ret+0(FP)

	MOVQ	$·irqHandler(SB), AX
	MOVQ	AX, ret+8(FP)

	RET

// func irq_enable()
TEXT ·irq_enable(SB),$0
	STI
	RET

// func irq_disable()
TEXT ·irq_disable(SB),$0
	CLI
	RET
