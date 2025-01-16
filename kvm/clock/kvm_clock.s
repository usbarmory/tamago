// KVM clock driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"

// holder for struct pvclock_vcpu_time_info
DATA	pvclock<>+0x00(SB)/8, $0x0000000000000000
DATA	pvclock<>+0x08(SB)/8, $0x0000000000000000
DATA	pvclock<>+0x10(SB)/8, $0x0000000000000000
GLOBL	pvclock<>(SB),8,$32

// func pvclock(msr uint32) uint32
TEXT ·pvclock(SB),$8
	MOVL	msr+0(FP), CX
	MOVL	$pvclock<>(SB), AX
	MOVL	$0, DX
	MOVL	AX, ret+8(FP)
	ORL	$1, AX
	WRMSR
	RET
