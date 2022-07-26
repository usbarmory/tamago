// ARM processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func irq_enable()
TEXT ·irq_enable(SB),$0
	WORD	$0xf10801c0 // cpsie aif
	RET

// func irq_disable()
TEXT ·irq_disable(SB),$0
	WORD	$0xf10c01c0 // cpsid aif
	RET
