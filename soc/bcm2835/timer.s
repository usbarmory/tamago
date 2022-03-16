// BCM2835 SoC support
// https://github.com/usbarmory/tamago
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func read_systimer() int64
TEXT ·read_systimer(SB),$0-8
	MOVW	·peripheralBase(SB), R2
	ADD	$0x00003000, R2 // timer peripheral offset

	MOVW	4(R2), R0 // lower 32-bits
	MOVW	8(R2), R1 // upper 32-bits

	MOVW	R0, ret_lo+0(FP)
	MOVW	R1, ret_hi+4(FP)

	RET
