// Nuvoton NUC980 early initialization
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// EarlyClockInit configures NUC980 clock gates from assembly.  Writes to
// the SYS/CLK register block (0xb0000000-0xb00003ff) silently fail
// when performed via the Go runtime's store path; direct STR from
// early assembly works reliably.

#include "textflag.h"

// func EarlyClockInit()
TEXT ·EarlyClockInit(SB),NOSPLIT|NOFRAME,$0
	// CLK_HCLKEN (0xb0000210): enable AIC AHB clock (bit 10).
	MOVW	$0xb0000210, R0
	MOVW	(R0), R1
	ORR	$(1<<10), R1
	MOVW	R1, (R0)

	// CLK_PCLKEN0 (0xb0000218): enable ETimer0 (bit 8),
	// ETimer1 (bit 9) and UART0 (bit 16) APB clocks.
	MOVW	$0xb0000218, R0
	MOVW	(R0), R1
	ORR	$(1<<8), R1
	ORR	$(1<<9), R1
	ORR	$(1<<16), R1
	MOVW	R1, (R0)

	// CLK_DIVCTL8 (0xb0000240): select XIN (12 MHz) for Timer0/Timer1
	// eclk source (bits [19:16] = 0).
	MOVW	$0xb0000240, R0
	MOVW	(R0), R1
	BIC	$(0xf<<16), R1
	MOVW	R1, (R0)

	// UART0 hardware init for pre-hwinit1 Printk (115200 8N1 from 12 MHz XIN).
	// UA_LCR (0xb007000c): 8-bit word, 1 stop, no parity.
	MOVW	$0xb007000c, R0
	MOVW	$0x3, R1
	MOVW	R1, (R0)
	// UA_FCR (0xb0070008): reset and enable TX/RX FIFOs.
	MOVW	$0xb0070008, R0
	MOVW	$0x6, R1
	MOVW	R1, (R0)
	// UA_BAUD (0xb0070024): Mode 2, BRD=0x66 → 115200 from 12 MHz XIN.
	MOVW	$0xb0070024, R0
	MOVW	$0x30000066, R1
	MOVW	R1, (R0)

	RET
