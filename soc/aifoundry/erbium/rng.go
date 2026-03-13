// AI Foundry Erbium RNG initialization
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package erbium

import (
	"encoding/binary"
	_ "unsafe"

	"github.com/usbarmory/tamago/internal/rng"
)

// FIXME
const seed uint64 = 0x1234567890abcdef

//go:linkname initRNG runtime/goos.InitRNG
func initRNG() {
	drbg := &rng.DRBG{}
	binary.LittleEndian.PutUint64(drbg.Seed[:], seed)
	rng.GetRandomDataFn = drbg.GetRandomData
}

// SetRNG allows to override the internal random number generator function used
// by TamaGo on the Erbium processor.
//
// At runtime initialization the erbium package seeds a DRBG with the CPU
// timer, as Erbium lacks an entropy source. This is unsuitable for secure
// random number generation and must therefore be overridden to ensure secure
// operation of Go crypto.
func SetRNG(getRandomData func([]byte)) {
	rng.GetRandomDataFn = getRandomData
}
