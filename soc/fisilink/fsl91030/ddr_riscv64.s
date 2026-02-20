// Fisilink FSL91030 DDR controller initialization
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

// func initDDR()
//
// DDR is expected to be pre-initialized by the boot loader before
// TamaGo starts. This stub is a no-op for that common case.
//
// If TamaGo is used as a first-stage boot loader and DDR init is required,
// implement the initialization sequence from freeloader.S here.
TEXT ·initDDR(SB),NOSPLIT|NOFRAME,$0
	RET
