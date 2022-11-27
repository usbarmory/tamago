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
)

const (
	m = 1 << 31
	a = 1103515245
	c = 12345
)

var lcg uint32

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

// GetLCGData implements a Linear Congruential Generator
// (https://en.wikipedia.org/wiki/Linear_congruential_generator)
//
// The LCG is used for random data on boards/SoCs which do not feature a valid
// entropy source.
//
// The LCG is unsuitable for secure random number generation and must therefore
// be overridden to ensure safe operation of Go `crypto/rand`.
func GetLCGData(b []byte) {
	if lcg == 0 {
		lcg = uint32(time.Now().UnixNano())
	}

	read := 0
	need := len(b)

	for read < need {
		lcg = (a*lcg + c) % m
		read = Fill(b, read, lcg)
	}
}
