// ARM processor support — ARMv5TEJ exception handlers
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// This file provides exception handlers for ARMv5TEJ cores (e.g. ARM926EJ-S)
// that lack the ARMv6+ MRS <reg>, <banked_reg> instruction used in
// exception.s.  It is selected automatically when GOARM=5 (the arm.6
// feature build tag is not set) and replaces exception.s.

//go:build !arm.6

#include "go_asm.h"
#include "textflag.h"

// set_exc_stack, set_vbar, set_mvbar use only standard ARM instructions
// and are identical to the default versions in exception.s.

// func set_exc_stack(addr uint32)
TEXT ·set_exc_stack(SB),NOSPLIT,$0-4
	MOVW addr+0(FP), R0

	// set FIQ mode SP
	WORD	$0xe321f0d1	// msr CPSR_c, #0xd1
	MOVW R0, R13

	// set IRQ mode SP
	WORD	$0xe321f0d2	// msr CPSR_c, #0xd2
	MOVW R0, R13

	// set Supervisor mode SP
	WORD	$0xe321f0d3	// msr CPSR_c, #0xd3
	MOVW R0, R13

	// set Monitor mode SP
	WORD	$0xe321f0d6	// msr CPSR_c, #0xd6
	MOVW R0, R13

	// set Abort mode SP
	WORD	$0xe321f0d7	// msr CPSR_c, #0xd7
	MOVW R0, R13

	// set Undefined mode SP
	WORD	$0xe321f0db	// msr CPSR_c, #0xdb
	MOVW R0, R13

	// return to System mode
	WORD	$0xe321f0df	// msr CPSR_c, #0xdf

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

// EXCEPTION_V5 is the ARMv5-compatible exception handler macro.
//
// On ARMv6+ exception.s uses "mrs sp, SP_usr" to load the user-mode
// stack pointer directly into the exception-mode SP.  This instruction
// does not exist on ARMv5TEJ.
//
// Instead, we:
//  0. Load the known exception stack address from the excStack global.
//     This is essential because a previous exception of the same type
//     will have left the banked exception SP pointing at the user stack.
//     The ARMv6+ version avoids this problem because "mrs sp, SP_usr"
//     atomically re-reads the user SP on every entry.
//  1. Save R0 and R1 to the exception-mode stack.
//  2. Record the exception stack address in R1.
//  3. Switch to SYS mode (which shares SP/LR with USR) via a fixed
//     immediate to CPSR_c.  R0 and R1 are not banked in any mode except
//     FIQ R8-R14, but we only use R0/R1 so this is safe even in FIQ.
//  4. Copy the user SP into R0.
//  5. Switch back to the original exception mode using the MODE parameter
//     (a fixed CPSR_c immediate such as 0xdb for UND).
//  6. Set SP to the user SP (from R0).
//  7. Restore R0 and R1 from the exception stack (addressed via R1).
//
// After this preamble the machine state is identical to the ARMv6+ version:
// SP = user stack pointer, all registers preserved.
//
// MODE is the CPSR_c immediate for the exception mode with I+F disabled:
//   FIQ=0xd1, IRQ=0xd2, SVC=0xd3, ABT=0xd7, UND=0xdb
#define EXCEPTION_V5(OFFSET, FN, LROFFSET, RN, SAVE_SIZE, MODE)	\
	/* reload exception SP from the excStack global so that    */	\
	/* repeated exceptions of the same type use a valid stack  */	\
	MOVW	·excStack(SB), R13					\
									\
	/* save R0, R1 to exception-mode stack */			\
	MOVM.DB.W	[R0-R1], (R13)		/* stmdb sp!, {r0,r1} */\
									\
	/* R1 = exception stack address (points to saved R0) */		\
	MOVW	R13, R1							\
									\
	/* switch to SYS mode (shares SP/LR with USR), I+F disabled */	\
	WORD	$0xe321f0df		/* msr CPSR_c, #0xdf */		\
									\
	/* R13 is now the user SP; copy it to R0 */			\
	MOVW	R13, R0							\
									\
	/* switch back to exception mode */				\
	WORD	$(0xe321f000 + MODE)	/* msr CPSR_c, #MODE */		\
									\
	/* set SP to user SP, restore R0/R1 from exception stack */	\
	MOVW	R0, R13			/* sp = user sp */		\
	WORD	$(0xe8910003)		/* ldmia r1, {r0, r1} */	\
									\
	/* --- from here identical to the ARMv6+ version --- */		\
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
	WORD	$0xe10f0000		/* mrs r0, CPSR */		\
	WORD	$0xe321f0db		/* msr CPSR_c, 0xdb */		\
	MOVW	g, R1							\
	WORD	$0xe129f000		/* msr CPSR, r0 */		\
	MOVW	R1, g							\
									\
	/* call exception handler on g0 */				\
	MOVW	$OFFSET, R0						\
	MOVW	$FN(SB), R1						\
	MOVW	$SAVE_SIZE, R2						\
	MOVW	R14, R3							\
	CALL	runtime·CallOnG0(SB)					\
									\
	/* restore caller registers */					\
	MOVM.IA.W	(R13), [R0-RN, R14]	/* pop {r0-rN, r14} */	\
									\
	/* restore PC from LR and mode */				\
	MOVW.S	R14, R15

// Reset and SVC both enter in SVC mode (0xd3).
TEXT ·resetHandler(SB),NOSPLIT|NOFRAME,$0
	EXCEPTION_V5(const_RESET, ·systemException, 0, R12, 56, 0xd3)

TEXT ·undefinedHandler(SB),NOSPLIT|NOFRAME,$0
	EXCEPTION_V5(const_UNDEFINED, ·systemException, 4, R12, 56, 0xdb)

TEXT ·supervisorHandler(SB),NOSPLIT|NOFRAME,$0
	EXCEPTION_V5(const_SUPERVISOR, ·systemException, 0, R12, 56, 0xd3)

TEXT ·prefetchAbortHandler(SB),NOSPLIT|NOFRAME,$0
	EXCEPTION_V5(const_PREFETCH_ABORT, ·systemException, 4, R12, 56, 0xd7)

TEXT ·dataAbortHandler(SB),NOSPLIT|NOFRAME,$0
	EXCEPTION_V5(const_DATA_ABORT, ·systemException, 8, R12, 56, 0xd7)

TEXT ·fiqHandler(SB),NOSPLIT|NOFRAME,$0
	EXCEPTION_V5(const_FIQ, ·systemException, 4, R7, 36, 0xd1)

TEXT ·nullHandler(SB),NOSPLIT|NOFRAME,$0
	MOVW.S	R14, R15
