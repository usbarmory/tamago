// LoongArch 64-bit processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

// The Go LoongArch assembler follows the Plan 9 destination-last operand
// convention, therefore for `rdtime.d rd, rj` and `cpucfg rd, rj` the result
// register (rd) is the last operand.

// func Rdtime() uint64
TEXT ·Rdtime(SB),NOSPLIT,$0-8
	RDTIMED	R4, R5		// R5 = stable counter value, R4 = counter id
	MOVV	R5, ret+0(FP)
	RET

// func read_cpucfg(sel uint64) uint64
TEXT ·read_cpucfg(SB),NOSPLIT,$0-16
	MOVV	sel+0(FP), R4
	CPUCFG	R4, R5		// R5 = cpucfg[R4]
	MOVV	R5, ret+8(FP)
	RET
