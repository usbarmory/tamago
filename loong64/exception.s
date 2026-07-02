// LoongArch 64-bit processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

// trapHandler is the common exception entry (ECFG.VS = 0), all exceptions and
// interrupts share this single vector and are dispatched in software.
//
// The default handler prints the cause and panics, therefore it never returns
// and no caller context needs to be preserved.
TEXT ·trapHandler(SB),NOSPLIT|NOFRAME,$0
	JMP	·systemException(SB)
