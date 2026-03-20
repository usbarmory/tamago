// Nuvoton NUC980 CRYPTO PRNG driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// CRYPTO PRNG driver for the NUC980 SoC.
//
// The NUC980 CRYPTO engine contains a PRNG that generates 256-bit (8×32-bit)
// keys.  This driver seeds the PRNG from a fixed seed and then re-triggers
// generation for each call to getRandomData.
//
// Register references: NUC980 Series Datasheet, p. 202 (§ 6.26 Cryptographic
// Accelerator).

package nuc980

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/internal/rng"
)

// CRYPTO register base and PRNG offsets
//
// NUC980 Series Datasheet, p. 202 (§ 6.26 Cryptographic Accelerator).
const (
	CRPT_BA = 0xB001C000

	CRPT_PRNG_CTL  = CRPT_BA + 0x008
	CRPT_PRNG_SEED = CRPT_BA + 0x00C
	CRPT_PRNG_KEY0 = CRPT_BA + 0x010
	CRPT_PRNG_KEY1 = CRPT_BA + 0x014
	CRPT_PRNG_KEY2 = CRPT_BA + 0x018
	CRPT_PRNG_KEY3 = CRPT_BA + 0x01C
	CRPT_PRNG_KEY4 = CRPT_BA + 0x020
	CRPT_PRNG_KEY5 = CRPT_BA + 0x024
	CRPT_PRNG_KEY6 = CRPT_BA + 0x028
	CRPT_PRNG_KEY7 = CRPT_BA + 0x02C
)

// PRNG_CTL register bits
const (
	PRNG_CTL_START    = 1 << 0 // Start PRNG generation
	PRNG_CTL_SEEDRLD  = 1 << 1 // Reload seed before generation
	PRNG_CTL_KEYSZ256 = 3 << 2 // Key size: 11 = 256-bit
	PRNG_CTL_BUSY     = 1 << 8 // PRNG busy bit (read-only)
)

// prngKeyRegs lists the eight 32-bit PRNG output key registers.
var prngKeyRegs = [8]uint32{
	CRPT_PRNG_KEY0,
	CRPT_PRNG_KEY1,
	CRPT_PRNG_KEY2,
	CRPT_PRNG_KEY3,
	CRPT_PRNG_KEY4,
	CRPT_PRNG_KEY5,
	CRPT_PRNG_KEY6,
	CRPT_PRNG_KEY7,
}

//go:linkname initRNG runtime/goos.InitRNG
func initRNG() {
	// Enable the CRPT AHB clock before the runtime can call GetRandomData.
	// InitRNG fires before Hwinit1, so the clock must be gated here rather
	// than in nuc980.Init() to avoid a busy-wait on an unclocked peripheral.
	reg.Or(REG_CLK_HCLKEN, HCLKEN_CRPT)

	rng.GetRandomDataFn = getRandomData
}

// prngSeeded tracks whether the PRNG has been seeded.  The initial seed
// is loaded once; subsequent calls let the PRNG advance its internal
// state without reloading, producing different output on every trigger.
var prngSeeded bool

// getRandomData fills b with bytes from the CRYPTO PRNG.
//
// Each call triggers a new 256-bit PRNG generation.  For requests larger
// than 32 bytes the generation is repeated as needed.
func getRandomData(b []byte) {
	read := 0
	need := len(b)

	for read < need {
		// KEYSZ bits [3:2]: 11 = 256-bit key.
		ctl := uint32(PRNG_CTL_KEYSZ256) | PRNG_CTL_START

		if !prngSeeded {
			// First call: load the seed and set SEEDRLD.  On
			// subsequent calls the PRNG advances its internal
			// state, producing a different sequence each time.
			//
			// TODO: replace 0xDEADBEEF with a genuine entropy
			// source (e.g. uninitialized SRAM noise, ADC
			// least-significant bits, or a boot-time counter
			// latched before DDR init) so that every boot
			// produces a different PRNG sequence.
			reg.Write(CRPT_PRNG_SEED, 0xDEADBEEF)
			ctl |= PRNG_CTL_SEEDRLD
			prngSeeded = true
		}

		reg.Write(CRPT_PRNG_CTL, ctl)

		// Wait for completion.
		for reg.Read(CRPT_PRNG_CTL)&PRNG_CTL_BUSY != 0 {
		}

		// Drain up to 32 bytes (8 × 32-bit words).
		for _, keyReg := range prngKeyRegs {
			if read >= need {
				break
			}
			read = rng.Fill(b, read, reg.Read(keyReg))
		}
	}
}
