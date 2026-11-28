// SpacemiT K1 RNG initialization
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package k1

import (
	"encoding/binary"
	"time"
	_ "unsafe"

	"github.com/usbarmory/tamago/internal/rng"
)

var drbg rng.DRBG

func getRandomData(b []byte) {
	drbg.GetRandomData(b)
}

//go:linkname initRNG runtime/goos.InitRNG
func initRNG() {
	binary.LittleEndian.PutUint64(drbg.Seed[:], uint64(time.Now().UnixNano()))
	rng.GetRandomDataFn = getRandomData
}

// SetRNG allows to override the internal random number generator function used
// by TamaGo on the K1 SoC.
//
// At runtime initialization the k1 package selects a timer seeded DRBG, which
// is unsuitable for secure random number generation and must therefore be
// overridden to ensure safe operation of Go `crypto/rand`.
// Note: K1 has TRNG, once it is implemented, the init should
// call SetRNG.
func SetRNG(getRandomData func([]byte)) {
	rng.GetRandomDataFn = getRandomData
}
