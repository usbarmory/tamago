// NXP Random Number Generator (RNGB) driver
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build tamago,arm

package imx6

import (
	"sync"
	_ "unsafe"

	"github.com/f-secure-foundry/tamago/internal/reg"
)

const (
	HW_RNG_BASE uint32 = 0x02284000

	HW_RNG_CMD    = HW_RNG_BASE + 0x04
	HW_RNG_CMD_SR = 6
	HW_RNG_CMD_CE = 5
	HW_RNG_CMD_GS = 1
	HW_RNG_CMD_ST = 0

	HW_RNG_CR    = HW_RNG_BASE + 0x08
	HW_RNG_CR_AR = 4

	HW_RNG_SR          = HW_RNG_BASE + 0x0c
	HW_RNG_SR_ST_PF    = 21
	HW_RNG_SR_ERR      = 16
	HW_RNG_SR_FIFO_LVL = 8
	HW_RNG_SR_SDN      = 5
	HW_RNG_SR_STDN     = 4

	HW_RNG_ESR = HW_RNG_BASE + 0x10
	HW_RNG_OUT = HW_RNG_BASE + 0x14
)

type rngb struct {
	sync.Mutex
}

var RNGB = &rngb{}

var lcg uint32
var getRandomDataFn func([]byte)

//go:linkname getRandomData runtime.getRandomData
func getRandomData(b []byte) {
	getRandomDataFn(b)
}

// getLCGData implements a Linear Congruential Generator
// (https://en.wikipedia.org/wiki/Linear_congruential_generator).
func getLCGData(b []byte) {
	if lcg == 0 {
		lcg = uint32(timerFn())
	}

	read := 0
	need := len(b)

	for read < need {
		lcg = (1103515245*lcg + 12345) % (1 << 31)
		read = fill(b, read, lcg)
	}
}

// Init initializes the RNGB module.
func (hw *rngb) Init() {
	hw.Lock()
	// note: cannot defer during initialization

	// p3105, 44.5.2 Automatic seeding, IMX6ULLRM

	// clear errors
	reg.Set(HW_RNG_CMD, HW_RNG_CMD_CE)

	// soft reset RNGB
	reg.Set(HW_RNG_CMD, HW_RNG_CMD_SR)

	// perform self-test
	reg.Set(HW_RNG_CMD, HW_RNG_CMD_ST)

	print("imx6_rng: self-test\n")
	for reg.Get(HW_RNG_SR, HW_RNG_SR_STDN, 0b1) != 1 {
		// reg.Wait cannot be used before runtime initialization
	}

	if reg.Get(HW_RNG_SR, HW_RNG_SR_ERR, 0b1) != 0 || reg.Get(HW_RNG_SR, HW_RNG_SR_ST_PF, 0b1) != 0 {
		panic("imx6_rng: self-test FAIL\n")
	}

	// enable auto-reseed
	reg.Set(HW_RNG_CR, HW_RNG_CR_AR)

	print("imx6_rng: seeding\n")
	for reg.Get(HW_RNG_SR, HW_RNG_SR_SDN, 0b1) != 1 {
		// reg.Wait cannot be used before runtime initialization
	}

	hw.Unlock()
}

func (hw *rngb) getRandomData(b []byte) {
	read := 0
	need := len(b)

	for read < need {
		if reg.Get(HW_RNG_SR, HW_RNG_SR_ERR, 0b1) != 0 {
			panic("imx6_rng: error during getRandomData\n")
		}

		if reg.Get(HW_RNG_SR, HW_RNG_SR_FIFO_LVL, 0b1111) > 0 {
			read = fill(b, read, reg.Read(HW_RNG_OUT))
		}
	}
}

func fill(b []byte, index int, val uint32) int {
	shift := 0
	limit := len(b)

	for (index < limit) && (shift <= 24) {
		b[index] = byte((val >> shift) & 0xff)
		index += 1
		shift += 8
	}

	return index
}
