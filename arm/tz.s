// ARM processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func read_scr() uint32
TEXT ·read_scr(SB),$0-4
	// ARM Architecture Reference Manual - ARMv7-A and ARMv7-R edition
	// B4.1.129 SCR, Secure Configuration Register, Security Extensions
	MRC	15, 0, R0, C1, C1, 0
	MOVW	R0, ret+0(FP)

	RET

// func write_nsacr(scr uint32)
TEXT ·write_nsacr(SB),$0-4
	// ARM Architecture Reference Manual - ARMv7-A and ARMv7-R edition
	// B4.1.111 NSACR, Non-Secure Access Control Register, Security Extensions
	MOVW	scr+0(FP), R0
	MCR	15, 0, R0, C1, C1, 2

	RET
