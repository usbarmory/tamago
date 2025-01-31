// KVM clock driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

#define KVM_HC_CLOCK_PAIRING		9
#define KVM_CLOCK_PAIRING_WALLCLOCK	0

// holder for struct pvclock_vcpu_time_info
DATA	pvclock<>+0x00(SB)/8, $0x0000000000000000
DATA	pvclock<>+0x08(SB)/8, $0x0000000000000000
DATA	pvclock<>+0x10(SB)/8, $0x0000000000000000
GLOBL	pvclock<>(SB),RODATA,$32

// func pvclock(msr uint32) uint32
TEXT ·pvclock(SB),$8
	MOVL	msr+0(FP), CX
	MOVL	$pvclock<>(SB), AX
	MOVL	$0, DX
	MOVL	AX, ret+8(FP)
	ORL	$1, AX
	WRMSR
	RET

// func kvmclock_pairing(ptr uint)
TEXT ·kvmclock_pairing(SB),$8
	MOVQ	$KVM_HC_CLOCK_PAIRING, AX
	MOVQ	ptr+0(FP), BX
	MOVQ	$KVM_CLOCK_PAIRING_WALLCLOCK, CX

	// vmcall
	BYTE	$0x0f
	BYTE	$0x01
	BYTE	$0xc1

	RET
