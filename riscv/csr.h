// RISC-V processor support
// https://github.com/usbarmory/tamago
//
// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in Go LICENSE file.

#define t0 5

#define CSRR(CSR,RD) WORD $(0x2073 + RD<<7 + CSR<<20)
#define CSRW(RS,CSR) WORD $(0x1073 + RS<<15 + CSR<<20)
