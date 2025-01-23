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

#define SYS_write		1
#define SYS_mmap		9
#define SYS_exit		60
#define SYS_clock_gettime	228
#define SYS_getrandom		318

TEXT cpuinit(SB),NOSPLIT|NOFRAME,$0
	MOVQ	runtime·ramStart(SB), DI
	MOVQ	runtime·ramSize(SB), SI
	MOVL	$0x3, DX	// PROT_READ | PROT_WRITE
	MOVL	$0x22, R10	// MAP_PRIVATE | MAP_ANONYMOUS
	MOVL	$0xffffffff, R8
	MOVL	$0, R9
	MOVL	$SYS_mmap, AX
	SYSCALL

	JMP	_rt0_tamago_start(SB)

// func sys_clock_gettime() int64
TEXT ·sys_clock_gettime(SB),NOSPLIT,$40-8
	SUBQ	$16, SP		// Space for results

	MOVL	$CLOCK_REALTIME, DI
	LEAQ	0(SP), SI
	MOVQ	$SYS_clock_gettime, AX
	SYSCALL

	MOVQ	0(SP), AX	// sec
	MOVQ	8(SP), DX	// nsec
	ADDQ	$16, SP

	IMULQ	$1000000000, AX
	ADDQ	DX, AX
	MOVQ	AX, ns+0(FP)

	RET

// func sys_exit(code int32)
TEXT ·sys_exit(SB), $0-4
	MOVL	code+0(FP), DI
	MOVL	$SYS_exit, AX
	SYSCALL
	RET

// func sys_write(c *byte)
TEXT ·sys_write(SB),NOSPLIT,$0-8
	MOVQ	$1, DI		// fd
	MOVQ	c+0(FP), SI	// p
	MOVL	$1, DX		// n
	MOVL	$SYS_write, AX
	SYSCALL
	RET

// func sys_getrandom(b []byte, n int)
TEXT ·sys_getrandom(SB), $0-32
	MOVQ	b+0(FP), DI
	MOVQ	n+24(FP), SI
	MOVL	$0, DX
	MOVL	$SYS_getrandom, AX
	SYSCALL
	RET
