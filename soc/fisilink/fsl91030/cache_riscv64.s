// Fisilink FSL91030 cache control (Nuclei UX600)
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

// Register numbers for CSR encoding
#define t0 5

// CSRS(RS,CSR): set bits in CSR using register RS (CSRRS x0, csr, rs)
#define CSRS(RS,CSR) WORD $(0x2073 + RS<<15 + CSR<<20)
// CSRC(RS,CSR): clear bits in CSR using register RS (CSRRC x0, csr, rs)
#define CSRC(RS,CSR) WORD $(0x3073 + RS<<15 + CSR<<20)

// CSR_MCACHE_CTL (Nuclei-specific): 0x7CA
#define mcachectl 0x7CA

// func enableCache()
TEXT ·enableCache(SB),NOSPLIT|NOFRAME,$0
	// Load CSR_CACHE_ENABLE (0x10001) into T0
	MOV	$0x10001, T0
	// csrs CSR_MCACHE_CTL, t0
	CSRS	(t0, mcachectl)
	RET

// func disableCache()
TEXT ·disableCache(SB),NOSPLIT|NOFRAME,$0
	// Load CSR_CACHE_ENABLE (0x10001) into T0
	MOV	$0x10001, T0
	// csrc CSR_MCACHE_CTL, t0
	CSRC	(t0, mcachectl)
	RET
