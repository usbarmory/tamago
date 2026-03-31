// AI Foundry Minion initialization
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build linkcpuinit

#include "go_asm.h"
#include "textflag.h"

TEXT cpuinit(SB),NOSPLIT|NOFRAME,$0
	MOV	$·ncpu(SB), A0
	MOV	$1, A2
	WORD	$0x02c5373b // AMOADDG.D a4,a2,(a0)

	// start additional hart
	CSRRS	ZERO, MHARTID, T0
	MOV	$0, T1
	BGT	T0, T1, apstart

	JMP	github·com∕usbarmory∕tamago∕riscv64·Init(SB)
apstart:
	JMP	·apstart(SB)
