// ARM processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

// func set_exc_stack(addr uint32)
TEXT ·set_exc_stack(SB),NOSPLIT,$0-4
	MOVW addr+0(FP), R0

	// Set FIQ mode SP
	WORD	$0xe321f0d1	// msr CPSR_c, 0xd1
	MOVW R0, R13

	// Set IRQ mode SP
	WORD	$0xe321f0d2	// msr CPSR_c, 0xd2
	MOVW R0, R13

	// Set Supervisor mode SP
	WORD	$0xe321f0d3	// msr CPSR_c, 0xd3
	MOVW R0, R13

	// Set Monitor mode SP
	WORD	$0xe321f0d6	// msr CPSR_c, 0xd6
	MOVW R0, R13

	// Set Abort mode SP
	WORD	$0xe321f0d7	// msr CPSR_c, 0xd7
	MOVW R0, R13

	// Set Undefined mode SP
	WORD	$0xe321f0db	// msr CPSR_c, 0xdb
	MOVW R0, R13

	// Return to System mode
	WORD	$0xe321f0df	// msr CPSR_c, 0xdf

	RET

// func set_vbar(addr uint32)
TEXT ·set_vbar(SB),NOSPLIT,$0-4
	MOVW	addr+0(FP), R0
	MCR	15, 0, R0, C12, C0, 0
	RET

// func set_mvbar(addr uint32)
TEXT ·set_mvbar(SB),NOSPLIT,$0-4
	MOVW	addr+0(FP), R0
	MCR	15, 0, R0, C12, C0, 1
	RET

#define EXCEPTION(OFFSET, FN, LROFFSET, RN, SAVE_SIZE)			\
	/* restore stack pointer */					\
	WORD	$0xe105d200			/* mrs sp, SP_usr */	\
									\
	/* remove exception specific LR offset */			\
	SUB	$LROFFSET, R14, R14					\
									\
	/* save caller registers */					\
	MOVM.DB.W	[R0-RN, R14], (R13)	/* push {r0-rN, r14} */	\
									\
	/* restore g in case this mode banks them */			\
	MOVW	$SAVE_SIZE, R0						\
	CMP	$44, R0							\
	B.GT	6(PC)							\
	WORD	$0xe10f0000			/* mrs r0, CPSR */	\
	WORD	$0xe321f0db			/* msr CPSR_c, 0xdb */	\
	MOVW	g, R1							\
	WORD	$0xe129f000			/* msr CPSR, r0 */	\
	MOVW	R1, g							\
									\
	/* call exception handler on g0 */				\
	MOVW	$OFFSET, R0						\
	MOVW	$FN(SB), R1						\
	MOVW	$SAVE_SIZE, R2						\
	MOVW	R14, R3							\
	CALL	runtime·CallOnG0(SB)					\
									\
	/* restore registers */						\
	MOVM.IA.W	(R13), [R0-RN, R14]	/* pop {r0-rN, r14} */	\
									\
	/* restore PC from LR and mode */				\
	MOVW.S	R14, R15

TEXT ·resetHandler(SB),NOSPLIT|NOFRAME,$0
	EXCEPTION(const_RESET, ·systemException, 0, R12, 56)

TEXT ·undefinedHandler(SB),NOSPLIT|NOFRAME,$0
	EXCEPTION(const_UNDEFINED, ·systemException, 4, R12, 56)

TEXT ·supervisorHandler(SB),NOSPLIT|NOFRAME,$0
	EXCEPTION(const_SUPERVISOR, ·systemException, 0, R12, 56)

TEXT ·prefetchAbortHandler(SB),NOSPLIT|NOFRAME,$0
	EXCEPTION(const_PREFETCH_ABORT, ·systemException, 4, R12, 56)

TEXT ·dataAbortHandler(SB),NOSPLIT|NOFRAME,$0
	EXCEPTION(const_DATA_ABORT, ·systemException, 8, R12, 56)

TEXT ·irqHandler(SB),NOSPLIT|NOFRAME,$0
	/* remove exception specific LR offset */
	SUB	$4, R14, R14

	/* save caller registers */
	MOVM.DB.W	[R0-R12, R14], (R13)	// push {r0-r12, r14}

	/* wake up IRQ handling goroutine */
	MOVW	·irqHandlerG(SB), R0
	MOVW	·irqHandlerP(SB), R1
	CMP	$0, R0
	B.EQ	done
	CMP	$0, R1
	B.EQ	done
	CALL	runtime·WakeG(SB)

	/* the IRQ handling goroutine is expected to unmask IRQs */
	WORD	$0xe14f0000			// mrs r0, SPSR
	ORR	$1<<7, R0			// mask IRQs
	WORD	$0xe169f000			// msr SPSR, r0
done:
	/* restore registers */
	MOVM.IA.W	(R13), [R0-R12, R14]	// pop {r0-r12, r14}

	/* restore PC from LR and mode */
	MOVW.S	R14, R15

TEXT ·fiqHandler(SB),NOSPLIT|NOFRAME,$0
	EXCEPTION(const_FIQ, ·systemException, 4, R7, 36)

TEXT ·nullHandler(SB),NOSPLIT|NOFRAME,$0
	MOVW.S	R14, R15
