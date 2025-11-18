// AMD64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func Fault()
TEXT ·Fault(SB),$0
	CLI

	// invalidate IDT
	MOVQ	$·idtptr(SB), AX
	MOVQ	$0, (AX)
	LIDT	(AX)

	// triple-fault
	CALL	$0
halt:
	HLT
	JMP halt

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
