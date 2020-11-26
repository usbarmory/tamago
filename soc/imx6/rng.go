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
	_ "unsafe"

	"github.com/f-secure-foundry/tamago/soc/imx6/rngb"
)

var lcg uint32
var getRandomDataFn func([]byte)

//go:linkname initRNG runtime.initRNG
func initRNG() {
	if Family == IMX6ULL && Native {
		rngb.Init()
		getRandomDataFn = rngb.GetRandomData
	} else {
		getRandomDataFn = getLCGData
	}
}

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
		read = rngb.Fill(b, read, lcg)
	}
}
