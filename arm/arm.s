// ARM processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func read_cpsr() uint32
TEXT ·read_cpsr(SB),$0-4
	// ARM Architecture Reference Manual - ARMv7-A and ARMv7-R edition
	// B1.3.3 Program Status Registers (PSRs)
	WORD	$0xe10f0000 // mrs r0, CPSR
	MOVW	R0, ret+0(FP)

	RET

// func halt()
TEXT ·halt(SB),$0
	// wait forever in low-power state
	WORD	$0xf10c0080 // cpsid i
	WORD	$0xe320f003 // wfi
