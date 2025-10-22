// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func read_cntfrq() uint32
TEXT ·read_cntfrq(SB),$0-4
	// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile
	// D12.8.1 CNTFRQ_EL0, Counter-timer Frequency register
	WORD	$0xdf3f03d5 // isb sy
	MRS	R0, CNTFRQ_EL0
	MOVW	R0, ret+0(FP)

	RET

// func write_cntfrq(freq uint32)
TEXT ·write_cntfrq(SB),$0-4
	// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile
	// D12.8.1 CNTFRQ_EL0, Counter-timer Frequency register
	MOVW	freq+0(FP), R0
	WORD	$0xdf3f03d5 // isb sy
	MSR	CNTFRQ_EL0, R0

	RET

// func write_cntkctl(val uint32)
TEXT ·write_cntkctl(SB),$0-4
	// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile
	// D12.8.15 CNTKCTL_EL1, Counter-timer Kernel Control register
	MOVW	val+0(FP), R0
	WORD	$0xdf3f03d5 // isb sy
	MSR	CNTKCTL_EL1, R0

	RET

// func read_cntpct() uint64
TEXT ·read_cntpct(SB),$0-8
	// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile
	// D12.8.19 CNTPCT_EL0, Counter-timer Physical Count register
	WORD	$0xdf3f03d5 // isb sy
	MRS	R0, CNTPCT_EL0
	MOVW	R0, ret+0(FP)

	RET
