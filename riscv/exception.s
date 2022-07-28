// RISC-V processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "csr.h"
#include "textflag.h"

#define mtvec  0x305
#define mepc   0x341
#define mcause 0x342

// func set_mtvec(addr uint64)
TEXT ·set_mtvec(SB),NOSPLIT,$0-8
	MOV	addr+0(FP), S0
	CSRW	(s0, mtvec)
	RET

// func read_mepc() uint64
TEXT ·read_mepc(SB),NOSPLIT,$0-8
	CSRR	(mepc, s0)
	MOV	S0, ret+0(FP)
	RET

// func read_mcause() uint64
TEXT ·read_mcause(SB),NOSPLIT,$0-8
	CSRR	(mcause, s0)
	MOV	S0, ret+0(FP)
	RET
