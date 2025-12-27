// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func ReadMSR(addr uint64) (val uint64)
TEXT ·ReadMSR(SB),$0-16
	MOVQ	addr+0(FP), CX
	RDMSR
	MOVL	AX, val+8(FP)
	MOVL	DX, val+4(FP)
	RET

// func WriteMSR(addr uint64, val uint64)
TEXT ·WriteMSR(SB),$0-16
	MOVQ	addr+0(FP), CX
	MOVL	val+8(FP),  AX
	MOVL	val+12(FP), DX
	WRMSR
	RET
