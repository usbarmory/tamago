// AMD64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

// Global Descriptor Table
DATA	·gdt+0x00(SB)/8, $0x0000000000000000	// null descriptor
DATA	·gdt+0x08(SB)/8, $0x00209a0000000000	// code descriptor (x/r)
DATA	·gdt+0x10(SB)/8, $0x0000920000000000	// data descriptor (r/w)
GLOBL	·gdt(SB),RODATA,$24

// Global Descriptor Table Register
DATA	·gdtptr+0x00(SB)/2, $(3*8-1)		// GDT Limit
DATA	·gdtptr+0x02(SB)/8, $·gdt(SB)		// GDT Base Address
GLOBL	·gdtptr(SB),RODATA,$(2+8)

// Interrupt Descriptor Table
GLOBL	·idt(SB),RODATA,$(const_vectors*16)

// Interrupt Descriptor Table Register
DATA	·idtptr+0x00(SB)/2, $(const_vectors*16-1)	// IDT Limit
DATA	·idtptr+0x02(SB)/8, $·idt(SB)			// IDT Base Address
GLOBL	·idtptr(SB),RODATA,$(2+8)

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
TEXT ·exit(SB),$0-4
	CLI
halt:
	HLT
	JMP halt

// func halt()
TEXT ·halt(SB),$0
	HLT
	RET
