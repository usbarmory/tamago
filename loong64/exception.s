// LoongArch 64-bit processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

// trapHandler is the common exception entry (ECFG.VS = 0); all exceptions and
// interrupts share this single vector and are dispatched in software on the
// ESTAT.Ecode field (0 identifies an interrupt).
TEXT ·trapHandler(SB),NOSPLIT|NOFRAME,$0
	WORD	$0x0400c024		// csrwr R4, SAVE0 (stash R4)
	WORD	$0x04001404		// csrrd R4, ESTAT
	SRLV	$16, R4, R4
	AND	$0x3f, R4, R4		// R4 = ESTAT.Ecode
	BNE	R4, R0, exception

	// interrupt
	WORD	$0x0400c004		// csrrd R4, SAVE0 (restore R4)
	JMP	·handleInterrupt(SB)

exception:
	WORD	$0x0400c004		// csrrd R4, SAVE0 (restore R4)
	JMP	·systemException(SB)

// handleInterrupt relays the interrupt to the os/signal subsystem and returns
// to the interrupted context with interrupts left masked (PRMD.PIE cleared), so
// the servicing goroutine is responsible for clearing the source and unmasking.
TEXT ·handleInterrupt(SB),NOSPLIT|NOFRAME,$0
	// save caller registers (R0 is zero, R3 is SP)
	MOVV	R1, -1*8(R3)
	MOVV	R2, -2*8(R3)
	MOVV	R4, -3*8(R3)
	MOVV	R5, -4*8(R3)
	MOVV	R6, -5*8(R3)
	MOVV	R7, -6*8(R3)
	MOVV	R8, -7*8(R3)
	MOVV	R9, -8*8(R3)
	MOVV	R10, -9*8(R3)
	MOVV	R11, -10*8(R3)
	MOVV	R12, -11*8(R3)
	MOVV	R13, -12*8(R3)
	MOVV	R14, -13*8(R3)
	MOVV	R15, -14*8(R3)
	MOVV	R16, -15*8(R3)
	MOVV	R17, -16*8(R3)
	MOVV	R18, -17*8(R3)
	MOVV	R19, -18*8(R3)
	MOVV	R20, -19*8(R3)
	MOVV	R21, -20*8(R3)
	MOVV	g,   -21*8(R3)
	MOVV	R23, -22*8(R3)
	MOVV	R24, -23*8(R3)
	MOVV	R25, -24*8(R3)
	MOVV	R26, -25*8(R3)
	MOVV	R27, -26*8(R3)
	MOVV	R28, -27*8(R3)
	MOVV	R29, -28*8(R3)
	MOVV	R30, -29*8(R3)
	MOVV	R31, -30*8(R3)

	ADDVU	$(-32*8), R3, R3
	MOVV	$(const_IRQ_SIGNAL), R4
	MOVV	R4, 8(R3)
	CALL	os∕signal·Relay(SB)
	ADDVU	$(32*8), R3, R3

	// keep interrupts masked on return: clear PRMD.PIE (bit 2)
	WORD	$0x04000404		// csrrd R4, PRMD
	MOVV	$4, R5
	ANDN	R5, R4, R4
	WORD	$0x04000424		// csrwr R4, PRMD

	// restore caller registers
	MOVV	-1*8(R3), R1
	MOVV	-2*8(R3), R2
	MOVV	-3*8(R3), R4
	MOVV	-4*8(R3), R5
	MOVV	-5*8(R3), R6
	MOVV	-6*8(R3), R7
	MOVV	-7*8(R3), R8
	MOVV	-8*8(R3), R9
	MOVV	-9*8(R3), R10
	MOVV	-10*8(R3), R11
	MOVV	-11*8(R3), R12
	MOVV	-12*8(R3), R13
	MOVV	-13*8(R3), R14
	MOVV	-14*8(R3), R15
	MOVV	-15*8(R3), R16
	MOVV	-16*8(R3), R17
	MOVV	-17*8(R3), R18
	MOVV	-18*8(R3), R19
	MOVV	-19*8(R3), R20
	MOVV	-20*8(R3), R21
	MOVV	-21*8(R3), g
	MOVV	-22*8(R3), R23
	MOVV	-23*8(R3), R24
	MOVV	-24*8(R3), R25
	MOVV	-25*8(R3), R26
	MOVV	-26*8(R3), R27
	MOVV	-27*8(R3), R28
	MOVV	-28*8(R3), R29
	MOVV	-29*8(R3), R30
	MOVV	-30*8(R3), R31

	// exception return
	WORD	$0x06483800		// ertn
