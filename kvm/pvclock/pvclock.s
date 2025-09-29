// KVM pvclock driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

// holder for struct pvclock_vcpu_time_info
DATA	·time_info<>+0x00(SB)/8, $0x0000000000000000
DATA	·time_info<>+0x08(SB)/8, $0x0000000000000000
DATA	·time_info<>+0x10(SB)/8, $0x0000000000000000
GLOBL	·time_info<>(SB),RODATA,$32

// func pvclock(msr uint32) uint32
TEXT ·pvclock(SB),$0-8
	MOVL	msr+0(FP), CX
	MOVL	$·time_info<>(SB), AX
	MOVL	$0, DX
	MOVL	AX, ret+8(FP)
	ORL	$1, AX
	WRMSR
	RET
