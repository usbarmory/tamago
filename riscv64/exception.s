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

#define stvec  0x105
#define sepc   0x141
#define scause 0x142

#define mtvec  0x305
#define mepc   0x341
#define mcause 0x342

// func set_stvec(addr uint64)
TEXT ·set_stvec(SB),NOSPLIT,$0-8
	MOV	addr+0(FP), T0
	CSRW	(t0, stvec)
	RET

// func read_sepc() uint64
TEXT ·read_sepc(SB),NOSPLIT,$0-8
	CSRR	(sepc, t0)
	MOV	T0, ret+0(FP)
	RET

// func read_scause() uint64
TEXT ·read_scause(SB),NOSPLIT,$0-8
	CSRR	(scause, t0)
	MOV	T0, ret+0(FP)
	RET

// func set_mtvec(addr uint64)
TEXT ·set_mtvec(SB),NOSPLIT,$0-8
	MOV	addr+0(FP), T0
	CSRW	(t0, mtvec)
	RET

// func read_mepc() uint64
TEXT ·read_mepc(SB),NOSPLIT,$0-8
	CSRR	(mepc, t0)
	MOV	T0, ret+0(FP)
	RET

// func read_mcause() uint64
TEXT ·read_mcause(SB),NOSPLIT,$0-8
	CSRR	(mcause, t0)
	MOV	T0, ret+0(FP)
	RET
