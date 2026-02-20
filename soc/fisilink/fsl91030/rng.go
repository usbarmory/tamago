// Fisilink FSL91030 RNG initialization
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package fsl91030

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
// by TamaGo on the FSL91030 SoC.
//
// At runtime initialization the fsl91030 package selects a timer seeded DRBG
// as the FSL91030 lacks a documented hardware entropy source. This is
// unsuitable for secure random number generation and must therefore be
// overridden to ensure safe operation of Go `crypto/rand`.
//
// If the FSL91030 has a hardware RNG peripheral, it should be configured and
// used here via SetRNG.
func SetRNG(getRandomData func([]byte)) {
	rng.GetRandomDataFn = getRandomData
}
