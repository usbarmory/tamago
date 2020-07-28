// ARM processor support
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func read_id_pfr0() uint32
TEXT ·read_idpfr0(SB),$0-4
	// ARM Architecture Reference Manual - ARMv7-A and ARMv7-R edition
	// https://wiki.osdev.org/ARMv7_Generic_Timers
	//
	// B4.1.93 ID_PFR0, Processor Feature Register 0, VMSA

	// Invalidate Entire Instruction Cache
	MOVW $0, R0
	MCR 15, 0, R0, C7, C5, 0

	MRC	15, 0, R0, C0, C1, 0

	MOVW	R0, ret+0(FP)

	RET

// func read_id_pfr1() uint32
TEXT ·read_idpfr1(SB),$0-4
	// ARM Architecture Reference Manual - ARMv7-A and ARMv7-R edition
	// https://wiki.osdev.org/ARMv7_Generic_Timers
	//
	// B4.1.94 ID_PFR1, Processor Feature Register 1, VMSA

	// Invalidate Entire Instruction Cache
	MOVW $0, R0
	MCR 15, 0, R0, C7, C5, 0

	MRC	15, 0, R0, C0, C1, 1

	MOVW	R0, ret+0(FP)

	RET
