// Linux user space support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

#define CLOCK_REALTIME 0

#define SYS_write		64
#define SYS_exit		93
#define SYS_clock_gettime	113
#define SYS_mmap		222
#define SYS_getrandom		278

TEXT cpuinit(SB),NOSPLIT|NOFRAME,$0
	MOVD	runtime∕goos·RamStart(SB), R0
	MOVD	runtime∕goos·RamSize(SB), R1
	MOVW	$0x3, R2	// PROT_READ | PROT_WRITE
	MOVW	$0x22, R3	// MAP_PRIVATE | MAP_ANONYMOUS
	MOVW	$0xffffffff, R4
	MOVW	$0, R5
	MOVW	$SYS_mmap, R8
	SVC

	// set stack pointer
	MOVD	runtime∕goos·RamStart(SB), R1
	MOVD	R1, RSP
	MOVD	runtime∕goos·RamSize(SB), R1
	MOVD	runtime∕goos·RamStackOffset(SB), R2
	ADD	R1, RSP
	SUB	R2, RSP

	B	_rt0_tamago_start(SB)

// func sys_clock_gettime() int64
TEXT ·sys_clock_gettime(SB),NOSPLIT,$40-8
	MOVD	RSP, R20
	MOVD	RSP, R1

	SUB	$16, R1
	BIC	$15, R1	// Align for C code
	MOVD	R1, RSP

	MOVW	$CLOCK_REALTIME, R0
	MOVD	$SYS_clock_gettime, R8
	SVC

	MOVD	0(RSP), R3	// sec
	MOVD	8(RSP), R5	// nsec

	MOVD	R20, RSP	// restore SP

	// sec is in R3, nsec in R5
	// return nsec in R3
	MOVD	$1000000000, R4
	MUL	R4, R3
	ADD	R5, R3
	MOVD	R3, ns+0(FP)
	RET

// func sys_exit(code int32)
TEXT ·sys_exit(SB), $0-4
	MOVW	code+0(FP), R0
	MOVD	$SYS_exit, R8
	SVC
	RET

// func sys_write(c *byte)
TEXT ·sys_write(SB),NOSPLIT,$0-8
	MOVW	$1, R0		// fd
	MOVD	c+0(FP), R1	// p
	MOVD	$1, R2		// n
	MOVW	$SYS_write, R8
	SVC
	RET

// func sys_getrandom(b []byte, n int)
TEXT ·sys_getrandom(SB), $0-32
	MOVD	b+0(FP), R0
	MOVD	n+24(FP), R1
	MOVW	$0, R2
	MOVW	$SYS_getrandom, R8
	SVC
	RET
