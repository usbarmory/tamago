// BCM2835 SoC Random Number Generator (RNG) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) the bcm2835 package authors
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package bcm2835

import (
	"sync"
	_ "unsafe"

	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/internal/rng"
)

// RNG registers
const (
	RNG_BASE = 0x104000

	RNG_CTRL = RNG_BASE + 0x0
	CTRL_EN  = 1

	RNG_STATUS = RNG_BASE + 0x4
	RNG_DATA   = RNG_BASE + 0x8
)

const warmupCount = 0x40000

// Rng represents a Random number generator instance
type Rng struct {
	sync.Mutex

	status uint32
	data   uint32
	ctrl   uint32
}

// RNG (Random Number Generator) instance
var RNG = &Rng{}

//go:linkname initRNG runtime.initRNG
func initRNG() {
	RNG.Init()
}

//go:linkname getRandomData runtime.getRandomData
func getRandomData(b []byte) {
	RNG.getRandomData(b)
}

// Init initializes the RNG by discarding 'warmup bytes'.
func (hw *Rng) Init() {
	hw.Lock()
	defer hw.Unlock()

	hw.status = PeripheralAddress(RNG_STATUS)
	hw.data = PeripheralAddress(RNG_DATA)
	hw.ctrl = PeripheralAddress(RNG_CTRL)

	// Discard
	reg.Write(hw.status, warmupCount)
	reg.Write(hw.ctrl, CTRL_EN)
}

func (hw *Rng) getRandomData(b []byte) {
	hw.Lock()
	defer hw.Unlock()

	read := 0
	need := len(b)

	for read < need {
		// Wait for at least one word to be available
		for (reg.Read(hw.status) >> 24) == 0 {
		}

		read = rng.Fill(b, read, reg.Read(hw.data))
	}
}
