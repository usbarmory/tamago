// LoongArch 64-bit processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

// func idle()
TEXT ·idle(SB),NOSPLIT|NOFRAME,$0
	WORD	$0x06488000	// idle 0
	RET
