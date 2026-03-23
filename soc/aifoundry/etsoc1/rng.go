// AI Foundry ET-SoC-1 RNG initialization
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package etsoc1

import (
	"encoding/binary"
	"time"
	_ "unsafe"

	"github.com/usbarmory/tamago/internal/rng"
)

//go:linkname initRNG runtime/goos.InitRNG
func initRNG() {
	drbg := &rng.DRBG{}
	binary.LittleEndian.PutUint64(drbg.Seed[:], uint64(time.Now().UnixNano()))
	rng.GetRandomDataFn = drbg.GetRandomData
}

// SetRNG allows to override the internal random number generator function used
// by TamaGo on the ET-SoC-1 processor.
//
// At runtime initialization the etsoc1 package seeds a DRBG with the CPU
// timer, as the ET-SoC-1 lacks an entropy source. This is unsuitable for
// secure random number generation and must therefore be overridden to ensure
// secure operation of Go crypto.
func SetRNG(getRandomData func([]byte)) {
	rng.GetRandomDataFn = getRandomData
}
