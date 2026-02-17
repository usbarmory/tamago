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
	"sync"

	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/internal/rng"
)

// TRNG registers
const (
	TRNG_CR   = 0x00
	CR_WAKEY  = 1
	CR_ENABLE = 0

	TRNG_MR  = 0x04
	MR_HALFR = 0

	TRNG_ISR   = 0x1ca
	ISR_DATRDY = 0

	TRNG_ODATA = 0x50

	WAKEY = 0x524e47 // "RNG"
)

// TRNG represents the TRNG instance.
type TRNG struct {
	sync.Mutex

	// Base register
	Base uint32
}

// Init initializes the TRNG module with automatic seeding.
func (hw *TRNG) Init() {
	hw.Lock()
	defer hw.Unlock()

	if hw.Base == 0 {
		panic("invalid TRNG instance")
	}

	// TODO: set only if clk > 100MHz
	reg.Set(hw.Base+TRNG_MR, MR_HALFR)

	// set Write Access Key and enable
	reg.Write(hw.Base+TRNG_CR, WAKEY|1)
}

// GetRandomData returns len(b) random bytes gathered from the RNGB module.
func (hw *TRNG) GetRandomData(b []byte) {
	read := 0
	need := len(b)

	for read < need {
		if reg.Get(hw.Base+TRNG_ISR, ISR_DATRDY, 1) > 0 {
			read = rng.Fill(b, read, reg.Read(hw.Base+TRNG_ODATA))
		}
	}
}
