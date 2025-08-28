// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func Msr(addr uint32) (val uint32)
TEXT ·Msr(SB),$0-12
	MOVL	addr+0(FP), CX
	RDMSR
	MOVL	AX, val+8(FP)
	RET
