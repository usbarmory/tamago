// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkcpuinit

#include "textflag.h"

TEXT cpuinit(SB),NOSPLIT|NOFRAME,$0
	B	_rt0_tamago_start(SB)

