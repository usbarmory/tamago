// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func In8(port uint16) (val uint8)
TEXT ·In8(SB),$0-9
	MOVW	port+0(FP), DX
	// in al, dx
	BYTE	$0xec
	MOVB	AL, val+8(FP)
	RET

// func Out8(port uint16, val uint8)
TEXT ·Out8(SB),$0-3
	MOVW	port+0(FP), DX
	MOVB	val+2(FP), AL
	// out dx, al
	BYTE	$0xee
	RET

// func In16(port uint16) (val uint16)
TEXT ·In16(SB),$0-10
	MOVW	port+0(FP), DX
	// in ax, dx
	BYTE	$0x66ed
	MOVW	AX, val+8(FP)
	RET

// func Out16(port uint16, val uint16)
TEXT ·Out16(SB),$0-4
	MOVW	port+0(FP), DX
	MOVW	val+2(FP), AX
	// out dx, al
	BYTE	$0x66ef
	RET

// func In32(port uint16) (val uint32)
TEXT ·In32(SB),$0-12
	MOVW	port+0(FP), DX
	// in eax, dx
	BYTE	$0xed
	MOVL	AX, val+8(FP)
	RET

// func Out32(port uint16, val uint32)
TEXT ·Out32(SB),$0-8
	MOVW	port+0(FP), DX
	MOVL	val+4(FP), AX
	// out dx, eax
	BYTE	$0xef
	RET
