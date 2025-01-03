// RISC-V processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// PMP CSRs helpers for RV64, only 8 PMPs are supported for now. In the future,
// to support up to 64 PMPs, this will benefit from dynamic generation with
// go:generate.

#include "csr.h"

#define pmpcfg0  0x3a0
#define pmpaddr0 0x3b0
#define pmpaddr1 0x3b1
#define pmpaddr2 0x3b2
#define pmpaddr3 0x3b3
#define pmpaddr4 0x3b4
#define pmpaddr5 0x3b5
#define pmpaddr6 0x3b6
#define pmpaddr7 0x3b7
#define pmpaddr8 0x3b8

// func read_pmpcfg0() uint64
TEXT ·read_pmpcfg0(SB),$0-8
	// Volume II: RISC-V Privileged Architectures V20211203
	// 3.7.1 Physical Memory Protection CSRs
	CSRR	(pmpcfg0, t0)
	MOV	T0, ret+0(FP)

	RET

// func write_pmpcfg0(uint64)
TEXT ·write_pmpcfg0(SB),$0-8
	// Volume II: RISC-V Privileged Architectures V20211203
	// 3.7.1 Physical Memory Protection CSRs
	MOV	cfg+0(FP), T0
	CSRW	(t0, pmpcfg0)

	RET

// func read_pmpaddr0() uint64
TEXT ·read_pmpaddr0(SB),$0-8
	CSRR	(pmpaddr0, t0)
	MOV	T0, ret+0(FP)
	RET

// func read_pmpaddr1() uint64
TEXT ·read_pmpaddr1(SB),$0-8
	CSRR	(pmpaddr1, t0)
	MOV	T0, ret+0(FP)
	RET

// func read_pmpaddr2() uint64
TEXT ·read_pmpaddr2(SB),$0-8
	CSRR	(pmpaddr2, t0)
	MOV	T0, ret+0(FP)
	RET

// func read_pmpaddr3() uint64
TEXT ·read_pmpaddr3(SB),$0-8
	CSRR	(pmpaddr3, t0)
	MOV	T0, ret+0(FP)
	RET

// func read_pmpaddr4() uint64
TEXT ·read_pmpaddr4(SB),$0-8
	CSRR	(pmpaddr4, t0)
	MOV	T0, ret+0(FP)
	RET

// func read_pmpaddr5() uint64
TEXT ·read_pmpaddr5(SB),$0-8
	CSRR	(pmpaddr5, t0)
	MOV	T0, ret+0(FP)
	RET

// func read_pmpaddr6() uint64
TEXT ·read_pmpaddr6(SB),$0-8
	CSRR	(pmpaddr6, t0)
	MOV	T0, ret+0(FP)
	RET

// func read_pmpaddr7() uint64
TEXT ·read_pmpaddr7(SB),$0-8
	CSRR	(pmpaddr7, t0)
	MOV	T0, ret+0(FP)
	RET

// func write_pmpaddr0(uint64)
TEXT ·write_pmpaddr0(SB),$0-8
	MOV	addr+0(FP), T0
	CSRW	(t0, pmpaddr0)
	RET

// func write_pmpaddr1(uint64)
TEXT ·write_pmpaddr1(SB),$0-8
	MOV	addr+0(FP), T0
	CSRW	(t0, pmpaddr1)
	RET

// func write_pmpaddr2(uint64)
TEXT ·write_pmpaddr2(SB),$0-8
	MOV	addr+0(FP), T0
	CSRW	(t0, pmpaddr2)
	RET

// func write_pmpaddr3(uint64)
TEXT ·write_pmpaddr3(SB),$0-8
	MOV	addr+0(FP), T0
	CSRW	(t0, pmpaddr3)
	RET

// func write_pmpaddr4(uint64)
TEXT ·write_pmpaddr4(SB),$0-8
	MOV	addr+0(FP), T0
	CSRW	(t0, pmpaddr4)
	RET

// func write_pmpaddr5(uint64)
TEXT ·write_pmpaddr5(SB),$0-8
	MOV	addr+0(FP), T0
	CSRW	(t0, pmpaddr5)
	RET

// func write_pmpaddr6(uint64)
TEXT ·write_pmpaddr6(SB),$0-8
	MOV	addr+0(FP), T0
	CSRW	(t0, pmpaddr6)
	RET

// func write_pmpaddr7(uint64)
TEXT ·write_pmpaddr7(SB),$0-8
	MOV	addr+0(FP), T0
	CSRW	(t0, pmpaddr7)
	RET
