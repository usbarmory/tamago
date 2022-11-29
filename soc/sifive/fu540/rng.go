// SiFive FU540 RNG initialization
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package fu540

import (
	"encoding/binary"
	"time"
	_ "unsafe"

	"github.com/usbarmory/tamago/internal/rng"
)

//go:linkname initRNG runtime.initRNG
func initRNG() {
	drbg := &rng.DRBG{}
	binary.LittleEndian.PutUint64(drbg.Seed[:], uint64(time.Now().UnixNano()))
	rng.GetRandomDataFn = drbg.GetRandomData

}

// SetRNG allows to override the internal random number generator function used
// by TamaGo on the FU540 SoC.
//
// At runtime initialization the fu540 package selects a timer seeded with the
// CPU timer as the FU540 lacks an entropy source. This is unsuitable for
// secure random number generation and must therefore be overridden to ensure
// safe operation of Go `crypto/rand`.
func SetRNG(getRandomData func([]byte)) {
	rng.GetRandomDataFn = getRandomData
}
