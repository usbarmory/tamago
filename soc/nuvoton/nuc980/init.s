// Nuvoton NUC980 SoC initialization
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

// EarlyClockInit gates the SoC clocks and brings up UART0 from assembly.
// Writes to the SYS/CLK register block via the Go runtime store path are
// unreliable this early, so direct stores from assembly are used instead.
TEXT ·EarlyClockInit(SB),NOSPLIT|NOFRAME,$0
	// enable AIC AHB clock
	MOVW	$const_REG_CLK_HCLKEN, R0
	MOVW	(R0), R1
	ORR	$const_HCLKEN_AIC, R1
	MOVW	R1, (R0)

	// enable ETimer0, ETimer1 and UART0 APB clocks
	MOVW	$const_REG_PCLKEN0, R0
	MOVW	(R0), R1
	ORR	$const_PCLKEN0_TMR0, R1
	ORR	$const_PCLKEN0_TMR1, R1
	ORR	$const_PCLKEN0_UA0, R1
	MOVW	R1, (R0)

	// select XIN (12 MHz) as the Timer0/Timer1 eclk source
	MOVW	$const_REG_CLK_DIV8, R0
	MOVW	(R0), R1
	BIC	$const_CLK_DIV8_ECLK_XIN, R1
	MOVW	R1, (R0)

	// UART0 8N1 @115200 from 12 MHz XIN for pre-Hwinit1 Printk
	MOVW	$(const_UA0_BA+0x0c), R0	// UA_LCR
	MOVW	$0x3, R1			// 8-bit word, 1 stop, no parity
	MOVW	R1, (R0)
	MOVW	$(const_UA0_BA+0x08), R0	// UA_FCR
	MOVW	$0x6, R1			// reset and enable TX/RX FIFOs
	MOVW	R1, (R0)
	MOVW	$(const_UA0_BA+0x24), R0	// UA_BAUD
	MOVW	$const_UA0_BAUD_115200, R1
	MOVW	R1, (R0)

	RET

// startup completes NUC980 early initialization after the board reset vector
// has applied any board specific pin multiplexing. It installs minimal
// exception vector stubs, gates the SoC clocks, clears BSS, sets up the stack,
// switches to System mode and enters the Go runtime.
TEXT ·startup(SB),NOSPLIT|NOFRAME,$0
	// Install minimal exception vectors (B . loops) so that any abort
	// during early init halts at a known address instead of cascading
	// through uninitialised DDR.
	MOVW	$0xeafffffe, R0
	MOVW	$0, R1
	MOVW	$8, R2
vecloop:
	MOVW	R0, (R1)
	ADD	$4, R1
	SUB	$1, R2
	CMP	$0, R2
	B.NE	vecloop

	// SoC clock gates (AIC, timers, UART0)
	BL	·EarlyClockInit(SB)

	// Zero BSS before reading Go variables (RamStart is in BSS and may
	// contain garbage if the DDR blob does not clear memory).
	BL	clearBSS(SB)

	// set stack pointer
	MOVW	runtime∕goos·RamStart(SB), R13
	MOVW	runtime∕goos·RamSize(SB), R1
	MOVW	runtime∕goos·RamStackOffset(SB), R2
	ADD	R1, R13
	SUB	R2, R13
	MOVW	R13, R3

	// detect HYP mode and switch to SVC if necessary
	WORD	$0xe10f0000	// mrs r0, CPSR
	AND	$0x1f, R0, R0

	CMP	$0x10, R0	// USR mode
	BL.EQ	_rt0_tamago_start(SB)

	CMP	$0x1a, R0	// HYP mode
	B.NE	after_eret

	BIC	$0x1f, R0
	ORR	$0x1d3, R0	// AIF masked, SVC mode
	MOVW	$12(R15), R14	// add lr, pc, #12 (after_eret)
	WORD	$0xe16ff000	// msr SPSR_fsxc, r0
	WORD	$0xe12ef30e	// msr ELR_hyp, lr
	WORD	$0xe160006e	// eret

after_eret:
	// enter System Mode
	WORD	$0xe321f0df	// msr CPSR_c, 0xdf

	MOVW	R3, R13
	B	_rt0_tamago_start(SB)

// clearBSS zeroes the .bss and .noptrbss sections.
TEXT clearBSS(SB),NOSPLIT|NOFRAME,$0
	MOVW	$runtime·bss(SB), R0
	MOVW	$runtime·enoptrbss(SB), R1
	MOVW	$0, R2
bss_loop:
	CMP	R0, R1
	B.EQ	bss_done
	MOVW	R2, (R0)
	ADD	$4, R0
	B	bss_loop
bss_done:
	RET
