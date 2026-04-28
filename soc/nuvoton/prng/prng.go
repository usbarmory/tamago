// Nuvoton CRYPTO PRNG driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package prng implements a driver for the pseudo random number generator
// found in the CRYPTO (Cryptographic Accelerator) engine of Nuvoton SoCs
// adopting the following reference specifications:
//   - NUC980 Series Datasheet - Rev 1.24
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package prng

import (
	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/internal/rng"
)

// CRYPTO PRNG register offsets (from PRNG.Base).
const (
	PRNG_CTL  = 0x008
	PRNG_SEED = 0x00c
	PRNG_KEY0 = 0x010
)

// PRNG_CTL register bits.
const (
	CTL_START    = 1 << 0 // start PRNG generation
	CTL_SEEDRLD  = 1 << 1 // reload seed before generation
	CTL_KEYSZ256 = 3 << 2 // key size: 0b11 = 256-bit
	CTL_BUSY     = 1 << 8 // PRNG busy (read-only)
)

// keyRegs is the number of 32-bit output key registers (256-bit / 32).
const keyRegs = 8

// PRNG represents a Nuvoton CRYPTO PRNG instance.
type PRNG struct {
	// Base register
	Base uint32
	// Seed is the value loaded on the first generation; subsequent calls
	// let the PRNG advance its internal state without reloading.
	Seed uint32

	seeded bool
}

// Init initializes the CRYPTO PRNG instance. The CRYPTO engine AHB clock
// must be gated on by the caller before use.
func (hw *PRNG) Init() {
	if hw.Base == 0 {
		panic("invalid PRNG instance")
	}
}

// GetRandomData fills b with bytes from the CRYPTO PRNG. Each generation
// produces 256 bits (8×32-bit); the generation is repeated as needed for
// requests larger than 32 bytes.
func (hw *PRNG) GetRandomData(b []byte) {
	read := 0
	need := len(b)

	for read < need {
		ctl := uint32(CTL_KEYSZ256) | CTL_START

		if !hw.seeded {
			reg.Write(hw.Base+PRNG_SEED, hw.Seed)
			ctl |= CTL_SEEDRLD
			hw.seeded = true
		}

		reg.Write(hw.Base+PRNG_CTL, ctl)

		// wait for completion
		for reg.Read(hw.Base+PRNG_CTL)&CTL_BUSY != 0 {
		}

		// drain up to 32 bytes (8×32-bit key words)
		for i := 0; i < keyRegs && read < need; i++ {
			read = rng.Fill(b, read, reg.Read(hw.Base+PRNG_KEY0+uint32(i*4)))
		}
	}
}
