// RISC-V 64-bit processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkcpuinit

#include "textflag.h"

#define t0 5

#define sie     0x104
#define mstatus 0x300
#define mie     0x304

#define CSRC(RS,CSR) WORD $(0x3073 + RS<<15 + CSR<<20)
#define CSRS(RS,CSR) WORD $(0x2073 + RS<<15 + CSR<<20)
#define CSRW(RS,CSR) WORD $(0x1073 + RS<<15 + CSR<<20)

TEXT cpuinit(SB),NOSPLIT|NOFRAME,$0
	// Disable interrupts
	MOV	$0, T0
	CSRW	(t0, sie)
	CSRW	(t0, mie)
	MOV	$0x7FFF, T0
	CSRC	(t0, mstatus)

	// Enable FPU
	MOV	$(1<<13), T0
	CSRS	(t0, mstatus)

	JMP	_rt0_tamago_start(SB)
