// ARM processor support
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func read_dbgauthstatus() uint32
TEXT ·read_dbgauthstatus(SB),$0-4
	// ARM Architecture Reference Manual - ARMv7-A and ARMv7-R edition
	//
	// C11.11.1 DBGAUTHSTATUS, Authentication Status register

	MRC	14, 0, R0, C7, C14, 6
	MOVW	R0, ret+0(FP)

	RET
