// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "arm64.h"

// func flush_tlb()
TEXT ·flush_tlb(SB),$0
	// invalidate TLBs
	DSB	SY
	ISB	SY

	// invalidate EL1 TLBs
	TLBI	VMALLE1
	DSB	SY
	ISB	SY

	RET

// func write_mair_el1(val uint64)
TEXT ·write_mair_el1(SB),$0-8
	// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile
	// D12.2.82 Memory Attribute Indirection Register (EL1)
	MOVD	val+0(FP), R0
	MSR	R0, MAIR_EL1
	ISB	SY

	RET

// func write_tcr_el1(val uint64)
TEXT ·write_tcr_el1(SB),$0-8
	// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile
	// D12.2.103 TCR_EL1, Translation Control Register (EL1)
	MOVD	val+0(FP), R0
	MSR	R0, TCR_EL1
	ISB	SY

	RET

// func set_ttbr0_el1(addr uint64)
TEXT ·set_ttbr0_el1(SB),$0-8
	// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile
	// D12.2.111 TTBR0_EL1, Translation Table Base Register 0 (EL1)
	MOVD	addr+0(FP), R0
	MSR	R0, TTBR0_EL1
	ISB	SY

	CALL	·flush_tlb(SB)

	MRS	SCTLR_EL1, R0
	BIC	$1<<19, R0	// clear WXN bit
	ORR	$1<<12, R0	// enable I-cache
	ORR	$1<<2, R0	// enable D-cache
	ORR	$1<<0, R0	// enable MMU
	MSR	R0, SCTLR_EL1
	ISB	SY

	RET
