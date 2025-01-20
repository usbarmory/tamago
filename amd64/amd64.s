// x86-64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func Fault()
TEXT ·Fault(SB),$0
	CLI
	XORL	AX, AX
	LGDT	(AX)
	HLT

// func halt(int32)
TEXT ·halt(SB),$0-8
	CLI
halt:
	HLT
	JMP halt
