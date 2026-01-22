// AMD Secure Encrypted Virtualization support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"

// func pvalidate(addr uint64, size int, validate bool) (ret uint32)
TEXT ·pvalidate(SB),$0-32
	MOVQ	addr+0(FP), AX
	MOVL	size+8(FP), CX
	MOVBQZX	validate+16(FP), DX

	// pvalidate
	BYTE	$0xf2
	BYTE	$0x0f
	BYTE	$0x01
	BYTE	$0xff

	MOVL AX, ret+24(FP)
	RET
