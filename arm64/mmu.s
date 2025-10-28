// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func write_mair_el3(val uint64)
TEXT ·write_mair_el3(SB),$0-8
	// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile
	// 4.3.69 Memory Attribute Indirection Register, EL1
	MOVD	val+0(FP), R0
	WORD	$0xd51ea200	// msr mair_el3, x0

	RET
