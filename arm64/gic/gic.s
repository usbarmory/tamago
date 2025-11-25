// ARM64 Generic Interrupt Controller (GICv3) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "../arm64.h"

// func write_icc_sre_el3(val uint64)
TEXT ·write_icc_sre_el3(SB),$0-8
	// ARM IHI 0069G
	// 12.2.24 ICC_SRE_EL3, Interrupt Controller System Register Enable register (EL3)
	MOVD	val+0(FP), R0
	WORD	$0xd51ecca0	// msr icc_sre_el3, x0
	ISB	SY

	RET

// func write_icc_igrpen0_el1(val uint64)
TEXT ·write_icc_igrpen0_el1(SB),$0-8
	// ARM IHI 0069G
	// 12.2.15 ICC_IGRPEN0_EL1, Interrupt Controller Interrupt Group 0 Enable register
	MOVD	val+0(FP), R0
	MSR	R0, ICC_IGRPEN0_EL1
	ISB	SY

	RET

// func write_icc_pmr_el1(val uint64)
TEXT ·write_icc_pmr_el1(SB),$0-8
	// ARM IHI 0069G
	// 12.2.18 ICC_PMR_EL1, Interrupt Controller Interrupt Priority Mask Register
	MOVD	val+0(FP), R0
	MSR	R0, ICC_PMR_EL1
	ISB	SY

	RET

// func read_mpidr_el1() uint64
TEXT ·read_mpidr_el1(SB),$0-8
	// ARM Cortex-A53 MPCore Processor Technical Reference Manual
	// 4.3.2 Multiprocessor Affinity Register
	ISB	SY
	MRS	MPIDR_EL1, R0
	MOVD	R0, ret+0(FP)

	RET

// func read_icc_iar0() uint64
TEXT ·read_icc_iar0(SB),$0-8
	// ARM IHI 0069G
	// 12.2.13 ICC_IAR0_EL1, Interrupt Controller Interrupt Acknowledge Register 0
	ISB	SY
	MRS	ICC_IAR0_EL1, R0
	MOVD	R0, ret+0(FP)

	RET

// func write_icc_eoir0(val uint64)
TEXT ·write_icc_eoir0(SB),$0-8
	// ARM IHI 0069G
	// 12.2.9 ICC_EOIR0_EL1, Interrupt Controller End Of Interrupt Register 0
	MOVD	val+0(FP), R0
	MSR	R0, ICC_EOIR0_EL1
	ISB	SY

	RET
