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
	SHLQ	$32, DX
	ORQ	DX, AX
	MOVQ	AX, val+8(FP)
	RET

// func WriteMSR(addr uint64, val uint64)
TEXT ·WriteMSR(SB),$0-16
	MOVQ	addr+0(FP), CX
	MOVQ	val+8(FP),  AX
	MOVQ	AX, DX
	SHRQ	$32, DX
	WRMSR
	RET
