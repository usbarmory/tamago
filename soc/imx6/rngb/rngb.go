// NXP Random Number Generator (RNGB) driver
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package rngb implements a driver for the NXP True Random Number Generator
// (RNGB) included in i.MX6ULL/i.MX6ULZ SoCs.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/f-secure-foundry/tamago.
package rngb

import (
	"sync"

	"github.com/f-secure-foundry/tamago/internal/reg"
)

// RNGB registers
const (
	RNG_BASE = 0x02284000

	RNG_CMD    = RNG_BASE + 0x04
	RNG_CMD_SR = 6
	RNG_CMD_CE = 5
	RNG_CMD_GS = 1
	RNG_CMD_ST = 0

	RNG_CR    = RNG_BASE + 0x08
	RNG_CR_AR = 4
	RNG_CR_GS = 1

	RNG_SR          = RNG_BASE + 0x0c
	RNG_SR_ST_PF    = 21
	RNG_SR_ERR      = 16
	RNG_SR_FIFO_LVL = 8
	RNG_SR_SDN      = 5
	RNG_SR_STDN     = 4

	RNG_ESR = RNG_BASE + 0x10
	RNG_OUT = RNG_BASE + 0x14
)

var mux sync.Mutex

// Reset resets the RNGB module.
func Reset() {
	mux.Lock()
	defer mux.Unlock()

	// soft reset RNGB
	reg.Set(RNG_CMD, RNG_CMD_SR)
}

// Init initializes the RNGB module with automatic seeding.
func Init() {
	mux.Lock()
	defer mux.Unlock()

	// p3105, 44.5.2 Automatic seeding, IMX6ULLRM

	// clear errors
	reg.Set(RNG_CMD, RNG_CMD_CE)

	// soft reset RNGB
	reg.Set(RNG_CMD, RNG_CMD_SR)

	// perform self-test
	reg.Set(RNG_CMD, RNG_CMD_ST)

	for reg.Get(RNG_SR, RNG_SR_STDN, 1) != 1 {
		// reg.Wait cannot be used before runtime initialization
	}

	if reg.Get(RNG_SR, RNG_SR_ERR, 1) != 0 || reg.Get(RNG_SR, RNG_SR_ST_PF, 1) != 0 {
		panic("imx6_rng: self-test failure\n")
	}

	// enable auto-reseed
	reg.Set(RNG_CR, RNG_CR_AR)
	// generate a seed
	reg.Set(RNG_CMD, RNG_CR_GS)

	for reg.Get(RNG_SR, RNG_SR_SDN, 1) != 1 {
		// reg.Wait cannot be used before runtime initialization
	}
}

// GetRandomData returns len(b) random bytes gathered from the RNGB module.
func GetRandomData(b []byte) {
	read := 0
	need := len(b)

	for read < need {
		if reg.Get(RNG_SR, RNG_SR_ERR, 1) != 0 {
			panic("imx6_rng: error during getRandomData\n")
		}

		if reg.Get(RNG_SR, RNG_SR_FIFO_LVL, 0b1111) > 0 {
			read = Fill(b, read, reg.Read(RNG_OUT))
		}
	}
}

func Fill(b []byte, index int, val uint32) int {
	shift := 0
	limit := len(b)

	for (index < limit) && (shift <= 24) {
		b[index] = byte((val >> shift) & 0xff)
		index += 1
		shift += 8
	}

	return index
}
