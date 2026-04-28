// ARM processor support — ARMv5TEJ exception handlers
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Exception handlers for ARMv5 cores that lack the ARMv6+
// "mrs sp, SP_usr" instruction used in exception.s.

//go:build !arm.6

#include "go_asm.h"
#include "textflag.h"

// set_exc_stack omits Monitor mode (ARMv6+ TrustZone only);
// set_vbar and set_mvbar are identical to exception.s.

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

	// Monitor mode (0xd6) is ARMv6+ TrustZone only; writing it
	// to CPSR on ARMv5 is UNPREDICTABLE.  Skipped.

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

// EXCEPTION_V5 is the ARMv5 exception handler macro.
//
// ARMv5 lacks "mrs sp, SP_usr" used by exception.s; instead the
// preamble loads excStack, saves R0/R1, round-trips through SYS mode
// to read the user SP, then switches back.  After this the state
// matches the ARMv6+ EXCEPTION macro.
//
// MODE is the CPSR_c immediate with I+F masked:
//   FIQ=0xd1, IRQ=0xd2, SVC=0xd3, ABT=0xd7, UND=0xdb
#define EXCEPTION_V5(OFFSET, FN, LROFFSET, RN, SAVE_SIZE, MODE)	\
	/* load exception SP (re-read on every entry) */		\
	MOVW	·excStack(SB), R13					\
									\
	/* save R0, R1 to exception stack */				\
	MOVM.DB.W	[R0-R1], (R13)		/* stmdb sp!, {r0,r1} */\
	MOVW	R13, R1							\
									\
	/* switch to SYS mode to read user SP */			\
	WORD	$0xe321f0df		/* msr CPSR_c, #0xdf */		\
	MOVW	R13, R0							\
									\
	/* switch back to exception mode */				\
	WORD	$(0xe321f000 + MODE)	/* msr CPSR_c, #MODE */		\
									\
	/* set SP to user SP, restore R0/R1 from exception stack */	\
	MOVW	R0, R13			/* sp = user sp */		\
	WORD	$(0xe8910003)		/* ldmia r1, {r0, r1} */	\
									\
	/* remainder identical to ARMv6+ EXCEPTION macro */		\
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
