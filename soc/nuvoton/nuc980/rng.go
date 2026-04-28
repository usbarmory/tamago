// Nuvoton NUC980 CRYPTO PRNG support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package nuc980

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/internal/rng"
	"github.com/usbarmory/tamago/soc/nuvoton/prng"
)

// CRPT_BA is the CRYPTO engine register base.
const CRPT_BA = 0xb001c000

// PRNG is the NUC980 CRYPTO engine pseudo random number generator.
//
// TODO: replace the fixed Seed with a genuine entropy source (e.g.
// uninitialized SRAM noise, ADC least-significant bits, or a boot-time
// counter latched before DDR init) so that every boot produces a
// different PRNG sequence.
var PRNG = &prng.PRNG{
	Base: CRPT_BA,
	Seed: 0xdeadbeef,
}

//go:linkname initRNG runtime/goos.InitRNG
func initRNG() {
	// Enable the CRPT AHB clock before the runtime can call GetRandomData.
	// InitRNG fires before Hwinit1, so the clock must be gated here rather
	// than in nuc980.Init() to avoid a busy-wait on an unclocked peripheral.
	reg.Or(REG_CLK_HCLKEN, HCLKEN_CRPT)

	PRNG.Init()
	rng.GetRandomDataFn = PRNG.GetRandomData
}
