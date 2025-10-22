// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func exit(int32)
TEXT ·exit(SB),$0-4
	// wait forever in low-power state
	MSR	DAIFSet, $0b1111	// disable all interrupts
	WORD	$0x7f2003d5		// wfi
