// Nuclei EvalSoC emulator support for tamago/riscv64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

// func rdtime() uint64
TEXT ·rdtime(SB),NOSPLIT,$0-8
	WORD	$0xc0102573	// rdtime a0 (csrr a0, time)
	MOV	X10, ret+0(FP)
	RET
