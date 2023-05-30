// ARM processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func read_gtc() int64
TEXT ·read_gtc(SB),$0-8
	// Cortex™-A9 MPCore® Technical Reference Manual
	// 4.4.1 Global Timer Counter Registers, 0x00 and 0x04
	//
	// p214, Table 2-1, ARM MP Global timer, IMX6DQRM
	MOVW	$0x00a00204, R1
	MOVW	$0x00a00200, R2
read:
	MOVW	(R1), R3
	MOVW	(R2), R4
	MOVW	(R1), R5
	CMP	R5, R3
	BNE	read

	MOVW	R3, ret_hi+4(FP)
	MOVW	R4, ret_lo+0(FP)

	RET

// func read_cntfrq() int32
TEXT ·read_cntfrq(SB),$0-4
	// ARM Architecture Reference Manual - ARMv7-A and ARMv7-R edition
	// B4.1.21 CNTFRQ, Counter Frequency register, VMSA
	WORD	$0xf57ff06f // isb sy
	MRC	15, 0, R0, C14, C0, 0

	MOVW	R0, ret+0(FP)

	RET

// func write_cntfrq(freq int32)
TEXT ·write_cntfrq(SB),$0-4
	// ARM Architecture Reference Manual - ARMv7-A and ARMv7-R edition
	// B4.1.21 CNTFRQ, Counter Frequency register, VMSA
	MOVW	freq+0(FP), R0

	WORD	$0xf57ff06f // isb sy
	MCR	15, 0, R0, C14, C0, 0

	RET

// func write_cntkctl(val uint32)
TEXT ·write_cntkctl(SB),$0-4
	// ARM Architecture Reference Manual - ARMv7-A and ARMv7-R edition
	// B4.1.26 CNTKCTL, Timer PL1 Control register, VMSA
	MOVW	val+0(FP), R0

	WORD	$0xf57ff06f // isb sy
	MCR	15, 0, R0, C14, C1, 0

	RET

// func read_cntpct() int64
TEXT ·read_cntpct(SB),$0-8
	// ARM Architecture Reference Manual - ARMv7-A and ARMv7-R edition
	// B4.1.30 CNTPCT, Physical Count register, VMSA
	WORD	$0xf57ff06f // isb sy
	WORD	$0xec510f0e // mrrc p15, 0, r0, r1, c14

	MOVW	R0, ret_lo+0(FP)
	MOVW	R1, ret_hi+4(FP)

	RET

// func write_cntptval(val int32, enable bool)
TEXT ·write_cntptval(SB),$0-8
	// ARM Architecture Reference Manual - ARMv7-A and ARMv7-R edition
	// B6.1.13 CNTP_TVAL, PL1 Physical TimerValue register, PMSA
	MOVW	val+0(FP), R0
	MOVW	enable+4(FP), R1

	WORD	$0xf57ff06f // isb sy
	MCR	15, 0, R0, C14, C2, 0
	MCR	15, 0, R1, C14, C2, 1

	RET

// func busyloop(count int32)
TEXT ·Busyloop(SB),$0-4
	MOVW count+0(FP), R0
loop:
	SUB.S	$1, R0, R0
	CMP	$0, R0
	BNE	loop

	RET
