// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "arm64.h"

// func read_cntfrq() uint32
TEXT ·read_cntfrq(SB),$0-4
	// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile
	// D12.8.1 CNTFRQ_EL0, Counter-timer Frequency register
	ISB	SY
	MRS	CNTFRQ_EL0, R0
	MOVW	R0, ret+0(FP)

	RET

// func write_cntfrq(freq uint32)
TEXT ·write_cntfrq(SB),$0-4
	// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile
	// D12.8.1 CNTFRQ_EL0, Counter-timer Frequency register
	MOVW	freq+0(FP), R0
	ISB	SY
	MSR	R0, CNTFRQ_EL0

	RET

// func write_cntkctl(val uint32)
TEXT ·write_cntkctl(SB),$0-4
	// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile
	// D12.8.15 CNTKCTL_EL1, Counter-timer Kernel Control register
	MOVW	val+0(FP), R0
	ISB	SY
	MSR	R0, CNTKCTL_EL1

	RET

// func read_cntpct() uint64
TEXT ·read_cntpct(SB),$0-8
	// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile
	// D12.8.19 CNTPCT_EL0, Counter-timer Physical Count register
	ISB	SY
	MRS	CNTPCT_EL0, R0
	MOVD	R0, ret+0(FP)

	RET

// func write_cntptval(val uint32, enable bool)
TEXT ·write_cntptval(SB),$0-8
	// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile
	// D12.8.18 CNTP_TVAL_EL0, Counter-timer Physical Timer TimerValue register
	MOVW	val+0(FP), R0
	MOVW	enable+4(FP), R1

	MSR	R0, CNTP_TVAL_EL0
	MSR	R1, CNTP_CTL_EL0

	RET
