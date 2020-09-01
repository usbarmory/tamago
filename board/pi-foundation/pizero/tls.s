// Raspberry Pi Zero Support
// https://github.com/f-secure-foundry/tamago
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

// For Pi Zero, implement runtime·read_tls_fallback
TEXT runtime·read_tls_fallback(SB),NOSPLIT|NOFRAME,$0
	WORD $0xee1d0f70 // mrc p15, 0, r0, c13, c0, 3
	RET
