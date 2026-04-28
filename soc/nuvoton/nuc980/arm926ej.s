// Nuvoton NUC980 SoC support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func readMIDR() uint32
//
// Read the Main ID Register (CP15 c0,c0,0).  Valid on all ARM cores
// from ARMv4 onward, including ARM926EJ-S.
// Reference: ARM Architecture Reference Manual, B3.12.36.
TEXT ·readMIDR(SB),$0-4
	MRC	15, 0, R0, C0, C0, 0
	MOVW	R0, ret+0(FP)
	RET

// func waitInterrupt()
//
// ARM926EJ-S lacks the ARMv7 WFI opcode; issue an equivalent CP15 operation:
//   MCR p15, 0, R0, c7, c0, 4  (stalls pipeline until interrupt or FIQ)
// Reference: ARM926EJ-S Technical Reference Manual, Section 2.3.8.
TEXT ·waitInterrupt(SB),$0
	WORD	$0xEE070F90 // MCR p15, 0, R0, c7, c0, 4
	RET

// func disableInterrupts()
//
// Masks both IRQ and FIQ; used by the SoC exit handler.
TEXT ·disableInterrupts(SB),$0
	WORD	$0xe10f0000 // MRS R0, CPSR
	ORR	$0xC0, R0
	WORD	$0xe129f000 // MSR CPSR_c, R0
	RET
