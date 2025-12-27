// KVM clock pairing driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

// https://docs.kernel.org/virt/kvm/x86/hypercalls.html
#define KVM_HC_CLOCK_PAIRING		9
#define KVM_CLOCK_PAIRING_WALLCLOCK	0

// holder for struct kvm_clock_pairing
GLOBL	·kvmclock<>(SB),RODATA,$8*3+4+4*9

// func Pairing() (sec int64, nsec int64, tsc uint64)
TEXT ·Pairing(SB),$0-24
	MOVQ	$KVM_HC_CLOCK_PAIRING, AX
	MOVQ	$·kvmclock<>(SB), BX
	MOVQ	$KVM_CLOCK_PAIRING_WALLCLOCK, CX

	// vmmcall
	BYTE	$0x0f
	BYTE	$0x01
	BYTE	$0xd9

	MOVQ	0(BX), AX
	MOVQ	AX, sec+0(FP)

	MOVQ	8(BX), AX
	MOVQ	AX, nsec+8(FP)

	MOVQ	16(BX), AX
	MOVQ	AX, tsc+16(FP)

	RET
