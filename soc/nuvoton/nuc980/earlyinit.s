// Nuvoton NUC980 early initialization
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// EarlyInit configures NUC980 clock gates from assembly.  Writes to
// the SYS/CLK register block (0xB0000000-0xB00003FF) silently fail
// when performed via the Go runtime's store path; direct STR from
// early assembly works reliably.

#include "textflag.h"

// func EarlyInit()
TEXT ·EarlyInit(SB),NOSPLIT|NOFRAME,$0
	// CLK_HCLKEN (0xB0000210): enable AIC AHB clock (bit 10).
	MOVW	$0xB0000210, R0
	MOVW	(R0), R1
	ORR	$(1<<10), R1
	MOVW	R1, (R0)

	// CLK_PCLKEN0 (0xB0000218): enable ETimer0 (bit 8),
	// ETimer1 (bit 9) and UART0 (bit 16) APB clocks.
	MOVW	$0xB0000218, R0
	MOVW	(R0), R1
	ORR	$(1<<8), R1
	ORR	$(1<<9), R1
	ORR	$(1<<16), R1
	MOVW	R1, (R0)

	// CLK_DIVCTL8 (0xB0000240): select XIN (12 MHz) for Timer0/Timer1
	// eclk source (bits [19:16] = 0).
	MOVW	$0xB0000240, R0
	MOVW	(R0), R1
	BIC	$(0xF<<16), R1
	MOVW	R1, (R0)

	RET
