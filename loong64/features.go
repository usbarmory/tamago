// LoongArch 64-bit processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package loong64

// CPUCFG word 2 feature bits.
const (
	cpucfg2FP     = 1 << 0  // basic floating point
	cpucfg2FPSP   = 1 << 1  // single precision floating point
	cpucfg2FPDP   = 1 << 2  // double precision floating point
	cpucfg2LSX    = 1 << 6  // 128-bit vector extension
	cpucfg2LASX   = 1 << 7  // 256-bit vector extension
	cpucfg2Crypto = 1 << 9  // cryptography extension
	cpucfg2LVZ    = 1 << 10 // virtualization extension
)

// Features represents the processor capabilities as reported by CPUCFG.
type Features struct {
	// FP indicates basic floating point support.
	FP bool
	// FPSP indicates single precision floating point support.
	FPSP bool
	// FPDP indicates double precision floating point support.
	FPDP bool
	// LSX indicates 128-bit vector (LSX) support.
	LSX bool
	// LASX indicates 256-bit vector (LASX) support.
	LASX bool
	// Crypto indicates cryptography acceleration support.
	Crypto bool
	// LVZ indicates hardware virtualization support.
	LVZ bool
}

// Features returns the processor capabilities.
func (cpu *CPU) Features() (features Features) {
	w2 := read_cpucfg(0x2)

	features.FP = w2&cpucfg2FP != 0
	features.FPSP = w2&cpucfg2FPSP != 0
	features.FPDP = w2&cpucfg2FPDP != 0
	features.LSX = w2&cpucfg2LSX != 0
	features.LASX = w2&cpucfg2LASX != 0
	features.Crypto = w2&cpucfg2Crypto != 0
	features.LVZ = w2&cpucfg2LVZ != 0

	return
}
