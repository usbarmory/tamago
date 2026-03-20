// Nuvoton NUC980 ETimer support
// https://github.com/usbarmory/tamago
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func readTimer() uint32
//
// Returns the current 24-bit value of the ETimer0 DR (Data Register).
// The timer runs at 1 MHz after initTimer(); each count = 1 µs.
TEXT ·readTimer(SB),$0-4
	MOVW	$0xB0050014, R0  // ETMR_DR = ETMR0_BA + 0x14
	MOVW	(R0), R1
	MOVW	R1, ret+0(FP)
	RET
