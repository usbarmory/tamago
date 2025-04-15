// x86-64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func Fault()
TEXT ·Fault(SB),$0
	CLI
	XORL	AX, AX
	LGDT	(AX)
	HLT

// func exit(int32)
TEXT ·exit(SB),$0-8
	CLI
halt:
	HLT
	JMP halt

// func halt()
TEXT ·halt(SB),$0
	HLT
	RET
