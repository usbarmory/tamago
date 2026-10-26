// uart
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func Read(addr uint32) uint32
TEXT ·Read(SB),$0-12
	MOVWU	addr+0(FP), T0
	MOVWU	(T0), T1
	MOVW	T1, ret+8(FP)

	RET

// func Read64(addr uint64) uint64
TEXT ·Read64(SB),$0-16
	MOV	addr+0(FP), T0
	MOV	(T0), T1
	MOV	T1, ret+8(FP)

	RET

// func Write(addr uint32, val uint32)
TEXT ·Write(SB),$0-8
	MOVWU	addr+0(FP), T0
	MOVWU	val+4(FP), T1
	MOVW	T1, (T0)

	RET
