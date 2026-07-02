// LoongArch 64-bit processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

// The Go LoongArch assembler provides no mnemonics for privileged CSR access,
// exception return (ertn) or the idle instruction, therefore these are emitted
// as hand-encoded WORD directives, verified via `go tool objdump`:
//
//	csrwr Rd, csr = 0x04000000 | (csr<<10) | (1<<5) | Rd
//	csrrd Rd, csr = 0x04000000 | (csr<<10) | Rd
//	idle  0       = 0x06488000
//
// CSR numbers: EUEN=0x2.

TEXT ·Init(SB),NOSPLIT|NOFRAME,$0
	// enable the floating point unit (EUEN.FPE)
	MOVV	$1, R4
	WORD	$0x04000824	// csrwr R4, EUEN(0x2)

	// set stack pointer to RamStart + RamSize - RamStackOffset
	MOVV	runtime∕goos·RamStart(SB), R3
	MOVV	runtime∕goos·RamSize(SB), R5
	MOVV	runtime∕goos·RamStackOffset(SB), R6
	ADDVU	R5, R3, R3
	SUBVU	R6, R3, R3

	JMP	_rt0_tamago_start(SB)

// func exit(int32)
TEXT ·exit(SB),NOSPLIT|NOFRAME,$0-4
	// wait forever in low-power state
again:
	WORD	$0x06488000	// idle 0
	JMP	again
