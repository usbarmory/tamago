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
	MRS	CPACR_EL1, R0
	ORR	$(3 << 20), R0	// set CPACR_EL1.FPEN
	MSR	R0, CPACR_EL1
	ISB	$1

	RET
