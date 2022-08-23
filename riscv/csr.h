// RISC-V processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#define s0 8

#define CSRR(CSR,RD) WORD $(0x2073 + RD<<7 + CSR<<20)
#define CSRW(RS,CSR) WORD $(0x1073 + RS<<15 + CSR<<20)
