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
	"github.com/usbarmory/tamago/bits"
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
	CTL_START   = 0 // start PRNG generation
	CTL_SEEDRLD = 1 // reload seed before generation
	CTL_KEYSZ   = 2 // key size field [3:2]
	CTL_BUSY    = 8 // PRNG busy (read-only)
)

// KEYSZ_256 selects a 256-bit key size in the CTL_KEYSZ field.
const KEYSZ_256 = 0b11

// keyRegs is the number of 32-bit output key registers (256-bit / 32).
const keyRegs = 8

// PRNG represents a Nuvoton CRYPTO PRNG instance.
type PRNG struct {
	// Base register
	Base uint32

	seed uint32
}

// Seed sets the value used to seed the CRYPTO PRNG. The CRYPTO PRNG is
// deterministic, so its output sequence is fully determined by this value.
// Seed can be called again at any time to re-seed a running PRNG.
func (hw *PRNG) Seed(seed uint32) {
	if seed == 0 {
		panic("RNG seed must not be 0")
	}

	var ctl uint32
	bits.SetN(&ctl, CTL_KEYSZ, 0b11, KEYSZ_256)
	bits.Set(&ctl, CTL_START)
	bits.Set(&ctl, CTL_SEEDRLD)

	reg.Write(hw.Base+PRNG_SEED, seed)
	reg.Write(hw.Base+PRNG_CTL, ctl)

	// wait for completion
	for reg.Get(hw.Base+PRNG_CTL, CTL_BUSY) {
	}

	hw.seed = seed
}

// GetRandomData fills b with bytes from the CRYPTO PRNG. Each generation
// produces 256 bits (8×32-bit); the generation is repeated as needed for
// requests larger than 32 bytes.
func (hw *PRNG) GetRandomData(b []byte) {
	if hw.seed == 0 {
		panic("RNG uninitialized")
	}

	read := 0
	need := len(b)

	for read < need {
		var ctl uint32
		bits.SetN(&ctl, CTL_KEYSZ, 0b11, KEYSZ_256)
		bits.Set(&ctl, CTL_START)
		reg.Write(hw.Base+PRNG_CTL, ctl)

		// wait for completion
		for reg.Get(hw.Base+PRNG_CTL, CTL_BUSY) {
		}

		// drain up to 32 bytes (8×32-bit key words)
		for i := 0; i < keyRegs && read < need; i++ {
			read = rng.Fill(b, read, reg.Read(hw.Base+PRNG_KEY0+uint32(i*4)))
		}
	}
}
