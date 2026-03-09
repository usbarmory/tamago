// RISC-V 64-bit processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "csr.h"
#include "textflag.h"

#define misa 0x301

// func read_misa() uint64
TEXT ·read_misa(SB),NOSPLIT,$0-8
	CSRR	(misa, t0)
	MOV	T0, ret+0(FP)
	RET
