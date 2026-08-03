// LoongArch 64-bit processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

// func read_crmd() uint64
TEXT ·read_crmd(SB),NOSPLIT,$0-8
	WORD	$0x04000004	// csrrd R4, CRMD
	MOVV	R4, ret+0(FP)
	RET

// func write_crmd(val uint64)
TEXT ·write_crmd(SB),NOSPLIT,$0-8
	MOVV	val+0(FP), R4
	WORD	$0x04000024	// csrwr R4, CRMD
	RET

// func read_ecfg() uint64
TEXT ·read_ecfg(SB),NOSPLIT,$0-8
	WORD	$0x04001004	// csrrd R4, ECFG
	MOVV	R4, ret+0(FP)
	RET

// func write_ecfg(val uint64)
TEXT ·write_ecfg(SB),NOSPLIT,$0-8
	MOVV	val+0(FP), R4
	WORD	$0x04001024	// csrwr R4, ECFG
	RET

// func read_estat() uint64
TEXT ·read_estat(SB),NOSPLIT,$0-8
	WORD	$0x04001404	// csrrd R4, ESTAT
	MOVV	R4, ret+0(FP)
	RET

// func read_era() uint64
TEXT ·read_era(SB),NOSPLIT,$0-8
	WORD	$0x04001804	// csrrd R4, ERA
	MOVV	R4, ret+0(FP)
	RET

// func read_badv() uint64
TEXT ·read_badv(SB),NOSPLIT,$0-8
	WORD	$0x04001c04	// csrrd R4, BADV
	MOVV	R4, ret+0(FP)
	RET

// func set_eentry(addr uint64)
TEXT ·set_eentry(SB),NOSPLIT,$0-8
	MOVV	addr+0(FP), R4
	WORD	$0x04003024	// csrwr R4, EENTRY
	RET

// func read_cpuid() uint64
TEXT ·read_cpuid(SB),NOSPLIT,$0-8
	WORD	$0x04008004	// csrrd R4, CPUID(0x20)
	MOVV	R4, ret+0(FP)
	RET

// func write_tcfg(val uint64)
TEXT ·write_tcfg(SB),NOSPLIT,$0-8
	MOVV	val+0(FP), R4
	WORD	$0x04010424	// csrwr R4, TCFG(0x41)
	RET

// func write_ticlr(val uint64)
TEXT ·write_ticlr(SB),NOSPLIT,$0-8
	MOVV	val+0(FP), R4
	WORD	$0x04011024	// csrwr R4, TICLR(0x44)
	RET
