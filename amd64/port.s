// x86-64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func Out(port uint32, val uint32)
TEXT ·Out(SB),$0-8
	MOVL	port+0(FP), DX
	MOVL	val+4(FP), AX
	// outb ax, dx
	BYTE	$0x66
	BYTE	$0xef
	RET
