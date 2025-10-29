// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "arm64.h"

// func write_mair_el3(val uint64)
TEXT ·write_mair_el3(SB),$0-8
	// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile
	// 4.3.69 Memory Attribute Indirection Register, EL1
	MOVD	val+0(FP), R0
	WORD	$0xd51ea200	// msr mair_el3, x0
	ISB	SY

	RET

// func write_tcr_el3(val uint32)
TEXT ·write_tcr_el3(SB),$0-4
	// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile
	// D12.2.105 TCR_EL3, Translation Control Register (EL3)
	MOVW	val+0(FP), R0
	WORD	$0xd51e2040	// msr tcr_el3, x0
	ISB	SY

	RET

// func set_ttbr0_el3(addr uint64)
TEXT ·set_ttbr0_el3(SB),$0-8
	// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile
	// D12.2.113 TTBR0_EL3, Translation Table Base Register 0 (EL3)
	MOVW	val+0(FP), R0
	WORD	$0xd51e2000	// msr ttbr0_el3, x0
	ISB	SY

	RET
