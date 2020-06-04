// NXP Random Number Generator (RNGB) driver
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package imx6

import (
	"sync"
	_ "unsafe"

	"github.com/f-secure-foundry/tamago/internal/reg"
)

// RNGB registers
const (
	RNG_BASE uint32 = 0x02284000

	RNG_CMD    = RNG_BASE + 0x04
	RNG_CMD_SR = 6
	RNG_CMD_CE = 5
	RNG_CMD_GS = 1
	RNG_CMD_ST = 0

	RNG_CR    = RNG_BASE + 0x08
	RNG_CR_AR = 4

	RNG_SR          = RNG_BASE + 0x0c
	RNG_SR_ST_PF    = 21
	RNG_SR_ERR      = 16
	RNG_SR_FIFO_LVL = 8
	RNG_SR_SDN      = 5
	RNG_SR_STDN     = 4

	RNG_ESR = RNG_BASE + 0x10
	RNG_OUT = RNG_BASE + 0x14
)

type Rng struct {
	sync.Mutex
}

// Random Number Generator (RNGB) instance
var RNGB = &Rng{}

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
		lcg = uint32(ARM.TimerFn())
	}

	read := 0
	need := len(b)

	for read < need {
		lcg = (1103515245*lcg + 12345) % (1 << 31)
		read = fill(b, read, lcg)
	}
}

// Init initializes the RNGB module.
func (hw *Rng) Init() {
	hw.Lock()
	// note: cannot defer during initialization

	// p3105, 44.5.2 Automatic seeding, IMX6ULLRM

	// clear errors
	reg.Set(RNG_CMD, RNG_CMD_CE)

	// soft reset RNGB
	reg.Set(RNG_CMD, RNG_CMD_SR)

	// perform self-test
	reg.Set(RNG_CMD, RNG_CMD_ST)

	print("imx6_rng: self-test\n")
	for reg.Get(RNG_SR, RNG_SR_STDN, 1) != 1 {
		// reg.Wait cannot be used before runtime initialization
	}

	if reg.Get(RNG_SR, RNG_SR_ERR, 1) != 0 || reg.Get(RNG_SR, RNG_SR_ST_PF, 1) != 0 {
		panic("imx6_rng: self-test FAIL\n")
	}

	// enable auto-reseed
	reg.Set(RNG_CR, RNG_CR_AR)

	print("imx6_rng: seeding\n")
	for reg.Get(RNG_SR, RNG_SR_SDN, 1) != 1 {
		// reg.Wait cannot be used before runtime initialization
	}

	hw.Unlock()
}

func (hw *Rng) getRandomData(b []byte) {
	read := 0
	need := len(b)

	for read < need {
		if reg.Get(RNG_SR, RNG_SR_ERR, 1) != 0 {
			panic("imx6_rng: error during getRandomData\n")
		}

		if reg.Get(RNG_SR, RNG_SR_FIFO_LVL, 0b1111) > 0 {
			read = fill(b, read, reg.Read(RNG_OUT))
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
