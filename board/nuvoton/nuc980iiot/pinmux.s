// NuMaker-IIoT-NUC980G2 board initialization
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build linkcpuinit

#include "go_asm.h"
#include "textflag.h"

// cpuinit applies board specific pin multiplexing and hands off to the NUC980
// SoC startup routine. It is the reset entry point selected by the linkcpuinit
// build tag (replacing the default arm cpuinit).
TEXT cpuinit(SB),NOSPLIT|NOFRAME,$0
	// Board pin mux: GPF11 = UART0_RXD, GPF12 = UART0_TXD.
	MOVW	$const_SYS_GPF_MFPH, R0
	MOVW	$const_GPF_MFPH_UART0, R1
	MOVW	R1, (R0)

	B	github·com∕usbarmory∕tamago∕soc∕nuvoton∕nuc980·startup(SB)
