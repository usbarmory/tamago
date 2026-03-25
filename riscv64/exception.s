// RISC-V 64-bit processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

// func set_stvec(addr uint64)
TEXT ·set_stvec(SB),NOSPLIT,$0-8
	MOV	addr+0(FP), T0
	CSRRW	ZERO, STVEC, T0
	RET

// func read_sepc() uint64
TEXT ·read_sepc(SB),NOSPLIT,$0-8
	CSRRS	ZERO, SEPC, T0
	MOV	T0, ret+0(FP)
	RET

// func read_scause() uint64
TEXT ·read_scause(SB),NOSPLIT,$0-8
	CSRRS	ZERO, SCAUSE, T0
	MOV	T0, ret+0(FP)
	RET

// func set_mtvec(addr uint64)
TEXT ·set_mtvec(SB),NOSPLIT,$0-8
	MOV	addr+0(FP), T0
	CSRRW	ZERO, MTVEC, T0
	RET

// func read_mepc() uint64
TEXT ·read_mepc(SB),NOSPLIT,$0-8
	CSRRS	ZERO, MEPC, T0
	MOV	T0, ret+0(FP)
	RET

// func read_mcause() uint64
TEXT ·read_mcause(SB),NOSPLIT,$0-8
	CSRRS	ZERO, MCAUSE, T0
	MOV	T0, ret+0(FP)
	RET
