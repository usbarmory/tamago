// AMD64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func rdrand() uint32
TEXT ·rdrand(SB),$0-4
	RDRANDL	AX
	MOVL	AX, ret+0(FP)
	RET
