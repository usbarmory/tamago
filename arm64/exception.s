// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

#define PAD					 	\
	WORD	$0xd503201f; WORD	$0xd503201f	\ // nop; nop
	WORD	$0xd503201f; WORD	$0xd503201f	\
	WORD	$0xd503201f; WORD	$0xd503201f	\
	WORD	$0xd503201f; WORD	$0xd503201f	\
	WORD	$0xd503201f; WORD	$0xd503201f	\
	WORD	$0xd503201f; WORD	$0xd503201f	\
	WORD	$0xd503201f; WORD	$0xd503201f	\
	WORD	$0xd503201f; WORD	$0xd503201f	\
	WORD	$0xd503201f; WORD	$0xd503201f	\
	WORD	$0xd503201f; WORD	$0xd503201f	\
	WORD	$0xd503201f; WORD	$0xd503201f	\
	WORD	$0xd503201f; WORD	$0xd503201f	\
	WORD	$0xd503201f; WORD	$0xd503201f	\
	WORD	$0xd503201f; WORD	$0xd503201f	\
	WORD	$0xd503201f; WORD	$0xd503201f	\
	WORD	$0xd503201f

// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile
// Table D1-7 Vector offsets from vector table base address
TEXT ·vectorTable(SB),NOSPLIT|NOFRAME,$0
	// EL0
	CALL	·systemException(SB); PAD // Synchronous Exception
	CALL	·systemException(SB); PAD // IRQ or vIRQ
	CALL	·systemException(SB); PAD // FIQ or vFIQ
	CALL	·systemException(SB); PAD // SError or vSError

	// ELx, x>0
	CALL	·systemException(SB); PAD // Synchronous Exception
	CALL	·systemException(SB); PAD // IRQ or vIRQ
	CALL	·systemException(SB); PAD // FIQ or vFIQ
	CALL	·systemException(SB); PAD // SError or vSError

// func set_vbar()
TEXT ·set_vbar(SB),NOSPLIT,$0
	MOVD	$·vectorTable(SB), R0
	WORD	$0xd51ec000	// msr vbar_el3, x0
	RET

// func read_el() uint64
TEXT ·read_el(SB),$0-8
	MRS	CurrentEL, R0
	MOVD	R0, ret+0(FP)

	RET
