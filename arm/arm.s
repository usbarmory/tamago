// ARM processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func read_cpsr() uint32
TEXT ·read_cpsr(SB),$0-4
	// ARM Architecture Reference Manual ARMv7-A and ARMv7-R edition
	// B1.3.3 Program Status Registers (PSRs)
	WORD	$0xe10f0000 // mrs r0, CPSR
	MOVW	R0, ret+0(FP)

	RET

// func exit(int32)
TEXT ·exit(SB),$0-4
	// wait forever in low-power state
	WORD	$0xe10f0000 // mrs r0, CPSR
	ORR	$1<<7, R0   // mask IRQs
	WORD	$0xe121f000 // msr CPSR_c, r0
	WORD	$0xe320f003 // wfi
