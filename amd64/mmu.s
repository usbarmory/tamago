// AMD64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func read_cr0() uint64
TEXT ·read_cr0(SB),$0-8
	MOVQ	CR0, AX
	MOVQ	AX, ret+0(FP)
	RET

// func write_cr0(val uint64)
TEXT ·write_cr0(SB),$0-8
	MOVQ	val+0(FP), AX
	MOVQ	AX, CR0

	RET

// func read_cr3() uint64
TEXT ·read_cr3(SB),$0-8
	MOVQ	CR3, AX
	MOVQ	AX, ret+0(FP)
	RET
