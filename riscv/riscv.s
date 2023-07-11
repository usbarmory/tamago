// RISC-V processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func halt()
TEXT ·halt(SB),$0
	// wait forever in low-power state
	WORD $0x10500073 // wfi
