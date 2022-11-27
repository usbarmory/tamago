// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// This code is adapted from Go `rand_plan9.go`.

package rng

import (
	"crypto/aes"
	"encoding/binary"
	"sync"
)

// DRBG is an AES-CTR based Deterministic Random Bit Generator. The generator
// is a fast key erasure RNG.
type DRBG struct {
	sync.Mutex

	// Seed represents the initial key for the AES-CTR cipher instance, it
	// will be overwritten during use to implement key erasure.
	Seed [32]byte
}

// GetRandomData returns len(b) random bytes.
func (r *DRBG) GetRandomData(b []byte) {
	var counter uint64
	var block [aes.BlockSize]byte

	r.Lock()
	blockCipher, err := aes.NewCipher(r.Seed[:])

	if err != nil {
		panic(err)
	}

	inc := func() {
		counter++
		if counter == 0 {
			panic("DRBG counter wrapped")
		}
		binary.LittleEndian.PutUint64(block[:], counter)
	}

	blockCipher.Encrypt(r.Seed[:aes.BlockSize], block[:])
	inc()
	blockCipher.Encrypt(r.Seed[aes.BlockSize:], block[:])
	inc()
	r.Unlock()

	for len(b) >= aes.BlockSize {
		blockCipher.Encrypt(b[:aes.BlockSize], block[:])
		inc()
		b = b[aes.BlockSize:]
	}
	if len(b) > 0 {
		blockCipher.Encrypt(block[:], block[:])
		copy(b, block[:])
	}
}
