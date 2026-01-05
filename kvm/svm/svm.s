// AMD secure virtualization support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#define PAGE_4K 0
#define PAGE_2M 1

// func vmgexit()
TEXT ·vmgexit(SB),$0
	BYTE	$0xf3
	BYTE	$0x0f
	BYTE	$0x01
	BYTE	$0xd9
	RET

// func pvalidate(addr uint64, validate bool) (ret uint32)
TEXT ·pvalidate(SB),$8
	MOVQ	addr+0(FP), AX
	MOVL	$PAGE_4K, CX
	MOVBLZX	validate+8(FP), DX

	// pvalidate
	BYTE	$0xf2
	BYTE	$0x0f
	BYTE	$0x01
	BYTE	$0xff

	MOVL AX, ret+16(FP)
	RET
