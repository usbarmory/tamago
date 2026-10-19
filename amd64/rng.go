// AMD64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package amd64

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/internal/rng"
)

// Intel® Digital Random Number Generator (DRNG)
// Software Implementation Guide - 5.2.1.
const rdrandRetries = 10

// defined in rng.s
func rdrand() (val uint64, ok bool)

func rdrandWord() uint64 {
	for i := 0; i < rdrandRetries; i++ {
		if val, ok := rdrand(); ok {
			return val
		}
	}

	panic("RDRAND failure")
}

// GetRandomData returns len(b) random bytes gathered from the RDRAND instruction.
func GetRandomData(b []byte) {
	read := 0
	need := len(b)

	for read < need {
		val := rdrandWord()
		read = rng.Fill(b, read, uint32(val))
		read = rng.Fill(b, read, uint32(val>>32))
	}
}

//go:linkname initRNG runtime/goos.InitRNG
func initRNG() {
	rng.GetRandomDataFn = GetRandomData
}
