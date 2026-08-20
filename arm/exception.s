// ARM processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

// func set_exc_stack(addr uint32)
TEXT ·set_exc_stack(SB),NOSPLIT,$0-4
	MOVW addr+0(FP), R0

	// set FIQ mode SP
	WORD	$0xe321f0d1	// msr CPSR_c, 0xd1
	MOVW R0, R13

	// set IRQ mode SP
	WORD	$0xe321f0d2	// msr CPSR_c, 0xd2
	MOVW R0, R13

	// set Supervisor mode SP
	WORD	$0xe321f0d3	// msr CPSR_c, 0xd3
	MOVW R0, R13
#ifndef GOARM_5
	// set Monitor mode SP
	WORD	$0xe321f0d6	// msr CPSR_c, 0xd6
	MOVW R0, R13
#endif
	// set Abort mode SP
	WORD	$0xe321f0d7	// msr CPSR_c, 0xd7
	MOVW R0, R13

	// set Undefined mode SP
	WORD	$0xe321f0db	// msr CPSR_c, 0xdb
	MOVW R0, R13

	// return to System mode
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

#define EXCEPTION(OFFSET, FN, LROFFSET)				\
	/* save exception vector */				\
	MOVW	$OFFSET, R0					\
	MOVW	R0, ·offset(SB)					\
	MOVW	R14, R0						\
								\
	/* remove exception specific LR offset */		\
	SUB	$LROFFSET, R14, R14				\
	MOVW	R14, ·eip(SB)					\
								\
	/* call exception handler on system stack (g0) */	\
	MOVW	FN(SB), R0					\
	MOVW	R0, 4(R13)					\
	CALL	runtime·CallOnG0(SB)

TEXT ·resetHandler(SB),NOSPLIT|NOFRAME,$0
	EXCEPTION(const_RESET, ·SystemExceptionHandler, 0)

TEXT ·undefinedHandler(SB),NOSPLIT|NOFRAME,$0
	EXCEPTION(const_UNDEFINED, ·SystemExceptionHandler, 4)

TEXT ·supervisorHandler(SB),NOSPLIT|NOFRAME,$0
	EXCEPTION(const_SUPERVISOR, ·SystemExceptionHandler, 0)

TEXT ·prefetchAbortHandler(SB),NOSPLIT|NOFRAME,$0
	EXCEPTION(const_PREFETCH_ABORT, ·SystemExceptionHandler, 4)

TEXT ·dataAbortHandler(SB),NOSPLIT|NOFRAME,$0
	EXCEPTION(const_DATA_ABORT, ·SystemExceptionHandler, 8)

TEXT ·fiqHandler(SB),NOSPLIT|NOFRAME,$0
	EXCEPTION(const_FIQ, ·SystemExceptionHandler, 4)

TEXT ·nullHandler(SB),NOSPLIT|NOFRAME,$0
	MOVW.S	R14, R15
