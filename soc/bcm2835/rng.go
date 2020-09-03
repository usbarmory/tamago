// BCM2835 SOC Random Number Generator
// https://github.com/f-secure-foundry/tamago
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package bcm2835

import (
	"sync"
	_ "unsafe"

	"github.com/f-secure-foundry/tamago/internal/reg"
)

const (
	// RNG_BASE is the offset within the peripheral address space
	RNG_BASE = 0x104000

	RNG_CTRL   = 0x0
	RNG_STATUS = 0x4
	RNG_DATA   = 0x8

	// RNG_RBGEN enables RNG
	RNG_RBGEN = 0x1
)

const warmupCount = 0x40000

var hwLock sync.Mutex

//go:linkname initRNG runtime.initRNG
func initRNG() {
	// note: cannot defer during initialization
	hwLock.Lock()

	// Discard
	reg.Write(PeripheralAddress(RNG_BASE+RNG_STATUS), warmupCount)
	reg.Write(PeripheralAddress(RNG_BASE+RNG_CTRL), RNG_RBGEN)

	hwLock.Unlock()
}

//go:linkname getRandomData runtime.getRandomData
func getRandomData(b []byte) {
	hwLock.Lock()

	read := 0
	need := len(b)

	for read < need {
		// Wait for at least one word to be available
		for (reg.Read(PeripheralAddress(RNG_BASE+RNG_STATUS)) >> 24) == 0 {
		}

		read = fill(b, read, reg.Read(PeripheralAddress(RNG_BASE+RNG_DATA)))
	}

	hwLock.Unlock()
}

func fill(b []byte, index int, val uint32) int {
	shift := 0
	limit := len(b)

	for (index < limit) && (shift <= 24) {
		b[index] = byte((val >> shift) & 0xff)
		index++
		shift += 8
	}

	return index
}
