// RISC-V processor support
// https://github.com/usbarmory/tamago
//
// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in Go LICENSE file.

#define t0 5

// CSRR(CSR,RD): rd = CSR (CSRRS with rs1=X0)
#define CSRR(CSR,RD) WORD $(0x2073 + RD<<7 + CSR<<20)
// CSRW(RS,CSR): CSR = rs1 (CSRRW with rd=X0)
#define CSRW(RS,CSR) WORD $(0x1073 + RS<<15 + CSR<<20)
// CSRS(RS,CSR): CSR |= rs1 (CSRRS with rd=X0)
#define CSRS(RS,CSR) WORD $(0x2073 + RS<<15 + CSR<<20)
// CSRC(RS,CSR): CSR &= ~rs1 (CSRRC with rd=X0)
#define CSRC(RS,CSR) WORD $(0x3073 + RS<<15 + CSR<<20)
