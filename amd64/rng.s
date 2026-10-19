// AMD64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func rdrand() (val uint64, ok bool)
TEXT ·rdrand(SB),$0-9
	RDRANDQ	AX
	SETCS	BX

	MOVQ	AX, val+0(FP)
	MOVB	BX, ok+8(FP)

	RET
