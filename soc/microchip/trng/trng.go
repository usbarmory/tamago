// Microchip True Random Number Generator (TRNG) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package trng implements a driver for the Microchip True Random Number Generator
// (TRNG) adopting the following specifications:
//   - Microchip - LAN9694/LAN9696/LAN9698 Datasheet - DS00005048E (02-27-25)
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm64` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package trng

import (
	"fmt"

	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/internal/rng"
)

const WAKEY = 0x524e47 // "RNG"

// TRNG registers
const (
	TRNG_CR   = 0x00
	CR_WAKEY  = 8
	CR_ENABLE = 0

	TRNG_MR  = 0x04
	MR_HALFR = 0

	TRNG_ISR   = 0x1c
	ISR_DATRDY = 0

	TRNG_ODATA = 0x50
	TRNG_WPSR  = 0xe8
)

// TRNG represents the TRNG instance.
type TRNG struct {
	// Base register
	Base uint32

	cycles int
}

// Init initializes the TRNG module with automatic seeding.
func (hw *TRNG) Init() {
	if hw.Base == 0 {
		panic("invalid TRNG instance")
	}

	// peripheral clk is 250MHz (> 100MHz)
	reg.Set(hw.Base+TRNG_MR, MR_HALFR)
	hw.cycles = 168

	// set Write Access Key and enable
	reg.Write(hw.Base+TRNG_CR, WAKEY<<CR_WAKEY|1)
}

// GetRandomData returns len(b) random bytes gathered from the TRNG module,
// [TRNG.Status] is set on error conditions.
func (hw *TRNG) GetRandomData(b []byte) {
	i := 0
	read := 0
	need := len(b)

	for read < need {
		if reg.GetN(hw.Base+TRNG_ISR, ISR_DATRDY, 1) > 0 {
			read = rng.Fill(b, read, reg.Read(hw.Base+TRNG_ODATA))
			i = 0
			continue
		}

		if isr := reg.Read(hw.Base + TRNG_ISR); isr > (1 << ISR_DATRDY) {
			panic(fmt.Sprintf("trng error, isr:%#x", isr))
		}

		if wpsr := reg.Read(hw.Base + TRNG_WPSR); wpsr > 0 {
			panic(fmt.Sprintf("trng error, wpsr:%#x", wpsr))
		}

		i++

		if i > hw.cycles {
			panic("trng unresponsive")
		}
	}
}
