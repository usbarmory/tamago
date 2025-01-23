// Linux user space support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

#define CLOCK_REALTIME 0

// for EABI, as we don't support OABI
#define SYS_BASE 0x0

#define SYS_exit		(SYS_BASE + 1)
#define SYS_write		(SYS_BASE + 4)
#define SYS_mmap2		(SYS_BASE + 192)
#define SYS_clock_gettime	(SYS_BASE + 263)
#define SYS_getrandom		(SYS_BASE + 384)

TEXT cpuinit(SB),NOSPLIT|NOFRAME,$0
	MOVW	runtime·ramStart(SB), R0
	MOVW	runtime·ramSize(SB), R1
	MOVW	$0x3, R2	// PROT_READ | PROT_WRITE
	MOVW	$0x22, R3	// MAP_PRIVATE | MAP_ANONYMOUS
	MOVW	$0xffffffff, R4
	MOVW	$0, R5
	MOVW	$SYS_mmap2, R7
	SWI	$0

	B	_rt0_tamago_start(SB)

// func sys_clock_gettime() int64
TEXT ·sys_clock_gettime(SB),NOSPLIT,$12-8
	MOVW	$CLOCK_REALTIME, R0
	MOVW	$spec-12(SP), R1	// timespec

	MOVW	$SYS_clock_gettime, R7
	SWI	$0

	MOVW	sec-12(SP), R0	// sec
	MOVW	nsec-8(SP), R2	// nsec

	MOVW	$1000000000, R3
	MULLU	R0, R3, (R1, R0)
	ADD.S	R2, R0
	ADC	$0, R1	// Add carry bit to upper half.

	MOVW	R0, ns_lo+0(FP)
	MOVW	R1, ns_hi+4(FP)

	RET

// func sys_exit(code int32)
TEXT ·sys_exit(SB), $0-4
	MOVW	code+0(FP), R0
	MOVW	$SYS_exit, R7
	SWI	$0
	RET

// func sys_write(c *byte)
TEXT ·sys_write(SB),NOSPLIT,$0-4
	MOVW	$1, R0		// fd
	MOVW	c+0(FP), R1	// p
	MOVW	$1, R2		// n
	MOVW	$SYS_write, R7
	SWI	$0
	RET

// func sys_getrandom(b []byte, n int)
TEXT ·sys_getrandom(SB), $0-16
	MOVW	b+0(FP), R0
	MOVW	n+12(FP), R1
	MOVW	$0, R2
	MOVW	$SYS_getrandom, R7
	SWI	$0
	RET
