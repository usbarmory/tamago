// ARM processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func read_actlr() uint32
TEXT ·read_actlr(SB),$0-4
	// Cortex™-A7 MPCore® Technical Reference Manual r0p5
	//
	// 4.3.31 Auxiliary Control Register

	// Invalidate Entire Instruction Cache
	MOVW	$0, R0
	MCR	15, 0, R0, C7, C5, 0

	MRC	15, 0, R0, C1, C0, 1
	MOVW	R0, ret+0(FP)

	RET

// func write_actlr(aux uint32)
TEXT ·write_actlr(SB),$0-4
	// Cortex™-A7 MPCore® Technical Reference Manual r0p5
	//
	// 4.3.31 Auxiliary Control Register

	// Invalidate Instruction Cache
	MOVW	$0, R1
	MCR	15, 0, R1, C7, C5, 0

	MOVW	aux+0(FP), R0
	MCR	15, 0, R0, C1, C0, 1

	RET

// func cache_disable()
TEXT ·cache_disable(SB),$0
	MRC	15, 0, R1, C1, C0, 0
	BIC	$1<<12, R1			// Disable I-cache
	BIC	$1<<2, R1			// Disable D-cache
	MCR	15, 0, R1, C1, C0, 0
	RET

// func cache_enable()
TEXT ·cache_enable(SB),$0
	MRC	15, 0, R1, C1, C0, 0
	ORR	$1<<12, R1			// Enable I-cache
	ORR	$1<<2, R1			// Enable D-cache
	MCR	15, 0, R1, C1, C0, 0
	RET

// Taken from Linux /arch/arm/mm/cache-v7.S
// Using R8 instead of R10 as the latter is g in go runtime.
//
// func cache_flush_data()
TEXT ·cache_flush_data(SB),$0
	WORD	$0xf57ff05f			// DMB SY
	MRC	15, 1, R0, C0, C0, 1		// read CLIDR
	MOVW	R0>>23, R3			// move LoC into position
	AND.S	$7<<1, R3, R3			// extract LoC*2 from clidr
	BEQ	finished			// if loc is 0, then no need to clean
start_flush_levels:
	MOVW	$0x0, R8			// start clean at cache level 0
flush_levels:
	ADD	R8>>1, R8, R2			// work out 3x current cache level
	MOVW	R0>>R2, R1			// extract cache type bits from clidr
	AND	$0x7, R1			// mask of the bits for current cache only
	CMP	$0x2, R1			// see what cache we have at this level
	BLT	skip				// skip if no cache, or just i-cache
	MCR	15, 2, R8, C0, C0, 0		// select current cache level in cssr
	WORD	$0xf57ff06f			// isb to sych the new cssr&csidr
	MRC	15, 1, R1, C0, C0, 0		// read the new csidr
	AND	$0x7, R1, R2			// extract the length of the cache lines
	ADD	$0x4, R2			// add 4 (line length offset)
	MOVW	$0x3ff, R4
	AND.S	R1>>3, R4, R4			// find maximum number on the way size
	CLZ	R4, R5				// find bit position of way size increment
	MOVW	$0x7fff, R7
	AND.S	R1>>13, R7, R7			// extract max number of the index size
loop1:
	MOVW	R7, R9				// create working copy of max index
loop2:
	ORR	R4<<R5, R8, R11			// factor way and cache number into r11
	ORR	R9<<R2, R11, R11		// factor way and cache number into r11
	MCR	15, 0, R11, C7, C14, 2		// clean & invalidate by set/way
	SUB.S	$1, R9, R9			// decrement the index
	BGE	loop2
	SUB.S	$1, R4, R4			// decrement the way
	BGE	loop1
skip:
	ADD	$2, R8				// increment cache number
	CMP	R8, R3
	//WORD	$0xf57ff04f			// DSB SY, CONFIG_ARM_ERRATA_814220, for Cortex-A7, not used in U-Boot
	BGT	flush_levels
finished:
	MOVW	$0, R8				// switch back to cache level 0
	MCR	15, 2, R8, C0, C0, 0		// select current cache level in cssr
	WORD	$0xf57ff04e			// DSB ST
	WORD	$0xf57ff06f			// ISB SY
	RET

// Taken from Linux /arch/arm/mm/cache-v7.S
// Using R8 instead of R10 as the latter is g in go runtime.
//
// func cache_flush_instruction()
TEXT ·cache_flush_instruction(SB),$0
	MOVW	$0, R0
	MCR	15, 0, R0, C7, C5, 0
	RET
