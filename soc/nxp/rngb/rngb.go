// NXP Random Number Generator (RNGB) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package rngb implements a driver for the NXP True Random Number Generator
// (RNGB) adopting the following specifications:
//   - IMX6ULLRM - i.MX 6ULL Applications Processor Reference Manual - Rev 1 2017/11
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/usbarmory/tamago.
package rngb

import (
	"sync"

	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/internal/rng"
)

// RNGB registers
const (
	RNG_CMD    = 0x04
	RNG_CMD_SR = 6
	RNG_CMD_CE = 5
	RNG_CMD_CI = 4
	RNG_CMD_GS = 1
	RNG_CMD_ST = 0

	RNG_CR    = 0x08
	RNG_CR_AR = 4
	RNG_CR_GS = 1

	RNG_SR          = 0x0c
	RNG_SR_ST_PF    = 21
	RNG_SR_ERR      = 16
	RNG_SR_FIFO_LVL = 8
	RNG_SR_SDN      = 5
	RNG_SR_STDN     = 4

	RNG_ESR = 0x10
	RNG_OUT = 0x14
)

// RNGB represents the RNGB instance.
type RNGB struct {
	sync.Mutex

	// Base register
	Base uint32

	// control registers
	cmd uint32
	cr  uint32
	sr  uint32
	esr uint32
	out uint32
}

// Reset resets the RNGB module.
func (hw *RNGB) Reset() {
	hw.Lock()
	defer hw.Unlock()

	// soft reset RNGB
	reg.Set(hw.cmd, RNG_CMD_SR)
}

// Init initializes the RNGB module with automatic seeding.
func (hw *RNGB) Init() {
	hw.Lock()
	defer hw.Unlock()

	if hw.Base == 0 {
		panic("invalid RNGB instance")
	}

	hw.cmd = hw.Base + RNG_CMD
	hw.cr = hw.Base + RNG_CR
	hw.sr = hw.Base + RNG_SR
	hw.esr = hw.Base + RNG_ESR
	hw.out = hw.Base + RNG_OUT

	// p3105, 44.5.2 Automatic seeding, IMX6ULLRM

	// clear errors
	reg.Set(hw.cmd, RNG_CMD_CE)

	// soft reset RNGB
	reg.Set(hw.cmd, RNG_CMD_SR)

	// perform self-test
	reg.Set(hw.cmd, RNG_CMD_ST)

	for reg.Get(hw.sr, RNG_SR_STDN, 1) != 1 {
		// reg.Wait cannot be used before runtime initialization
	}

	if reg.Get(hw.sr, RNG_SR_ERR, 1) != 0 || reg.Get(hw.sr, RNG_SR_ST_PF, 1) != 0 {
		panic("rngb: self-test failure\n")
	}

	// enable auto-reseed
	reg.Set(hw.cr, RNG_CR_AR)
	// generate a seed
	reg.Set(hw.cmd, RNG_CR_GS)

	for reg.Get(hw.sr, RNG_SR_SDN, 1) != 1 {
		// reg.Wait cannot be used before runtime initialization
	}

	// clear interrupts
	reg.Set(hw.cmd, RNG_CMD_CI)
}

// GetRandomData returns len(b) random bytes gathered from the RNGB module.
func (hw *RNGB) GetRandomData(b []byte) {
	read := 0
	need := len(b)

	for read < need {
		if reg.Get(hw.sr, RNG_SR_ERR, 1) != 0 {
			panic("rngb: error\n")
		}

		if reg.Get(hw.sr, RNG_SR_FIFO_LVL, 0b1111) > 0 {
			read = rng.Fill(b, read, reg.Read(hw.out))
		}
	}
}
