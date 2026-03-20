// Nuvoton NUC980 SoC support
// https://github.com/usbarmory/tamago
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func waitInterrupt()
//
// ARM926EJ-S lacks the ARMv7 WFI opcode; issue an equivalent CP15 operation:
//   MCR p15, 0, R0, c7, c0, 4  (stalls pipeline until interrupt or FIQ)
// Reference: ARM926EJ-S Technical Reference Manual, Section 2.3.8.
TEXT ·waitInterrupt(SB),$0
	WORD	$0xEE070F90 // MCR p15, 0, R0, c7, c0, 4
	RET

// func disableInterrupts()
TEXT ·disableInterrupts(SB),$0
	WORD	$0xe10f0000 // MRS R0, CPSR
	ORR	$0xC0, R0
	WORD	$0xe129f000 // MSR CPSR_c, R0
	RET

// func irqEnableV5(spsr bool)
TEXT ·irqEnableV5(SB),$0
	CMP	$1, R0
	B.EQ	irqen_spsr

	WORD	$0xe10f0000 // MRS R0, CPSR
	BIC	$0x80, R0
	WORD	$0xe129f000 // MSR CPSR_c, R0
	RET
irqen_spsr:
	WORD	$0xe14f0000 // MRS R0, SPSR
	BIC	$0x80, R0
	WORD	$0xe169f000 // MSR SPSR_c, R0
	RET

// func irqDisableV5(spsr bool)
TEXT ·irqDisableV5(SB),$0
	CMP	$1, R0
	B.EQ	irqdis_spsr

	WORD	$0xe10f0000 // MRS R0, CPSR
	ORR	$0x80, R0
	WORD	$0xe129f000 // MSR CPSR_c, R0
	RET
irqdis_spsr:
	WORD	$0xe14f0000 // MRS R0, SPSR
	ORR	$0x80, R0
	WORD	$0xe169f000 // MSR SPSR_c, R0
	RET

// func fiqEnableV5(spsr bool)
TEXT ·fiqEnableV5(SB),$0
	CMP	$1, R0
	B.EQ	fiqen_spsr

	WORD	$0xe10f0000 // MRS R0, CPSR
	BIC	$0x40, R0
	WORD	$0xe129f000 // MSR CPSR_c, R0
	RET
fiqen_spsr:
	WORD	$0xe14f0000 // MRS R0, SPSR
	BIC	$0x40, R0
	WORD	$0xe169f000 // MSR SPSR_c, R0
	RET

// func fiqDisableV5(spsr bool)
TEXT ·fiqDisableV5(SB),$0
	CMP	$1, R0
	B.EQ	fiqdis_spsr

	WORD	$0xe10f0000 // MRS R0, CPSR
	ORR	$0x40, R0
	WORD	$0xe129f000 // MSR CPSR_c, R0
	RET
fiqdis_spsr:
	WORD	$0xe14f0000 // MRS R0, SPSR
	ORR	$0x40, R0
	WORD	$0xe169f000 // MSR SPSR_c, R0
	RET
