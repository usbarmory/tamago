// ARM processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func irq_enable(spsr bool)
TEXT ·irq_enable(SB),$0
	CMP	$1, R0
	B.EQ	spsr

	WORD	$0xf1080080 // cpsie i
	RET
spsr:
	WORD	$0xe14f0000 // mrs r0, SPSR
	BIC	$1<<7, R0   // unmask IRQs
	WORD	$0xe169f000 // msr SPSR, r0
	RET

// func irq_disable(spsr bool)
TEXT ·irq_disable(SB),$0
	CMP	$1, R0
	B.EQ	spsr

	WORD	$0xf10c0080 // cpsid i
	RET
spsr:
	WORD	$0xe14f0000 // mrs r0, SPSR
	ORR	$1<<7, R0   // mask IRQs
	WORD	$0xe169f000 // msr SPSR, r0
	RET

// func fiq_enable(spsr bool)
TEXT ·fiq_enable(SB),$0
	CMP	$1, R0
	B.EQ	spsr

	WORD	$0xf1080040 // cpsie f
	RET
spsr:
	WORD	$0xe14f0000 // mrs r0, SPSR
	BIC	$1<<6, R0   // unmask FIQs
	WORD	$0xe169f000 // msr SPSR, r0
	RET

// func fiq_disable(spsr bool)
TEXT ·fiq_disable(SB),$0
	CMP	$1, R0
	B.EQ	spsr

	WORD	$0xf10c0040 // cpsid f
	RET
spsr:
	WORD	$0xe14f0000 // mrs r0, SPSR
	ORR	$1<<6, R0   // mask FIQs
	WORD	$0xe169f000 // msr SPSR, r0
	RET
