// SpacemiT K1 boot vector (linkcpuinit)
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// This file overrides the default riscv64 cpuinit when built with the
// 'linkcpuinit' build tag.
//
// The K1 DRAM, cache and PLL initialization is performed by the boot ROM
// loaded first stage (e.g. U-Boot SPL), which is also expected to keep the
// secondary harts parked. What such stage does not necessarily leave behind is
// a machine mode context suitable for the Go runtime, which requires the FPU
// as the compiler emits F/D instructions.
//
// The vector below therefore only normalizes the machine mode context before
// the runtime starts: interrupts are disabled and the FPU is enabled
// (MSTATUS.FS = Initial), the remaining startup (stack setup) is left to the
// runtime as in the default riscv64 cpuinit.

//go:build linkcpuinit

#include "textflag.h"

// MSTATUS[14:0] covers the interrupt-enable, SPP, VS, MPP and FS fields
#define MSTATUS_LOW_MASK   0x7fff
// MSTATUS.FS field low bit, 0b01 = Initial (FPU on)
#define MSTATUS_FS_INITIAL 13

TEXT cpuinit(SB),NOSPLIT|NOFRAME,$0
	// disable machine and supervisor interrupts
	MOV	$0, T0
	CSRRW	T0, SIE, ZERO
	CSRRW	T0, MIE, ZERO

	// clear MSTATUS[14:0] (interrupt-enable and FS bits)
	MOV	$MSTATUS_LOW_MASK, T0
	CSRRC	T0, MSTATUS, ZERO

	// then set FS = 0b01 (Initial) to enable the FPU
	MOV	$(1<<MSTATUS_FS_INITIAL), T0
	CSRRS	T0, MSTATUS, ZERO

	JMP	github·com∕usbarmory∕tamago∕riscv64·Init(SB)
