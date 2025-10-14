// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"

// func fp_enable()
TEXT ·fp_enable(SB),$0
	MRS	R0, CPACR_EL1
	ORR	R0, R0, $(3 << 20)	// set CPACR_EL1.FPEN
	MRS	CPACR_EL1, R0
	ISB

	RET
