// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package rng

import (
	"time"
	_ "unsafe"
)

var lcg uint32
var GetRandomDataFn func([]byte)

//go:linkname getRandomData runtime.getRandomData
func getRandomData(b []byte) {
	GetRandomDataFn(b)
}

// getLCGData implements a Linear Congruential Generator
// (https://en.wikipedia.org/wiki/Linear_congruential_generator).
func GetLCGData(b []byte) {
	if lcg == 0 {
		lcg = uint32(time.Now().UnixNano())
	}

	read := 0
	need := len(b)

	for read < need {
		lcg = (1103515245*lcg + 12345) % (1 << 31)
		read = Fill(b, read, lcg)
	}
}

func Fill(b []byte, index int, val uint32) int {
	shift := 0
	limit := len(b)

	for (index < limit) && (shift <= 24) {
		b[index] = byte((val >> shift) & 0xff)
		index += 1
		shift += 8
	}

	return index
}
