// RISC-V 64-bit processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// PMP CSRs helpers for RV64, only 8 PMPs are supported for now. In the future,
// to support up to 64 PMPs, this will benefit from dynamic generation with
// go:generate.

// func read_pmpcfg0() uint64
TEXT ·read_pmpcfg0(SB),$0-8
	// Volume II: RISC-V Privileged Architectures V20211203
	// 3.7.1 Physical Memory Protection CSRs
	CSRRS	ZERO, PMPCFG0, T0
	MOV	T0, ret+0(FP)

	RET

// func write_pmpcfg0(uint64)
TEXT ·write_pmpcfg0(SB),$0-8
	// Volume II: RISC-V Privileged Architectures V20211203
	// 3.7.1 Physical Memory Protection CSRs
	MOV	cfg+0(FP), T0
	CSRRW	T0, PMPCFG0, ZERO

	RET

// func read_pmpaddr0() uint64
TEXT ·read_pmpaddr0(SB),$0-8
	CSRRS	ZERO, PMPADDR0, T0
	MOV	T0, ret+0(FP)
	RET

// func read_pmpaddr1() uint64
TEXT ·read_pmpaddr1(SB),$0-8
	CSRRS	ZERO, PMPADDR1, T0
	MOV	T0, ret+0(FP)
	RET

// func read_pmpaddr2() uint64
TEXT ·read_pmpaddr2(SB),$0-8
	CSRRS	ZERO, PMPADDR2, T0
	MOV	T0, ret+0(FP)
	RET

// func read_pmpaddr3() uint64
TEXT ·read_pmpaddr3(SB),$0-8
	CSRRS	ZERO, PMPADDR3, T0
	MOV	T0, ret+0(FP)
	RET

// func read_pmpaddr4() uint64
TEXT ·read_pmpaddr4(SB),$0-8
	CSRRS	ZERO, PMPADDR4, T0
	MOV	T0, ret+0(FP)
	RET

// func read_pmpaddr5() uint64
TEXT ·read_pmpaddr5(SB),$0-8
	CSRRS	ZERO, PMPADDR5, T0
	MOV	T0, ret+0(FP)
	RET

// func read_pmpaddr6() uint64
TEXT ·read_pmpaddr6(SB),$0-8
	CSRRS	ZERO, PMPADDR6, T0
	MOV	T0, ret+0(FP)
	RET

// func read_pmpaddr7() uint64
TEXT ·read_pmpaddr7(SB),$0-8
	CSRRS	ZERO, PMPADDR7, T0
	MOV	T0, ret+0(FP)
	RET

// func write_pmpaddr0(uint64)
TEXT ·write_pmpaddr0(SB),$0-8
	MOV	addr+0(FP), T0
	CSRRW	T0, PMPADDR0, ZERO
	RET

// func write_pmpaddr1(uint64)
TEXT ·write_pmpaddr1(SB),$0-8
	MOV	addr+0(FP), T0
	CSRRW	T0, PMPADDR1, ZERO
	RET

// func write_pmpaddr2(uint64)
TEXT ·write_pmpaddr2(SB),$0-8
	MOV	addr+0(FP), T0
	CSRRW	T0, PMPADDR2, ZERO
	RET

// func write_pmpaddr3(uint64)
TEXT ·write_pmpaddr3(SB),$0-8
	MOV	addr+0(FP), T0
	CSRRW	T0, PMPADDR3, ZERO
	RET

// func write_pmpaddr4(uint64)
TEXT ·write_pmpaddr4(SB),$0-8
	MOV	addr+0(FP), T0
	CSRRW	T0, PMPADDR4, ZERO
	RET

// func write_pmpaddr5(uint64)
TEXT ·write_pmpaddr5(SB),$0-8
	MOV	addr+0(FP), T0
	CSRRW	T0, PMPADDR5, ZERO
	RET

// func write_pmpaddr6(uint64)
TEXT ·write_pmpaddr6(SB),$0-8
	MOV	addr+0(FP), T0
	CSRRW	T0, PMPADDR6, ZERO
	RET

// func write_pmpaddr7(uint64)
TEXT ·write_pmpaddr7(SB),$0-8
	MOV	addr+0(FP), T0
	CSRRW	T0, PMPADDR7, ZERO
	RET
