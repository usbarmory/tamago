// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Adapted from go/src/crypto/ecdsa/ecdsa_test.go

package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"time"
)

func testSignAndVerify(c elliptic.Curve, tag string) {
	start := time.Now()
	fmt.Printf("ECDSA sign and verify with p%d ... ", c.Params().BitSize)

	priv, _ := ecdsa.GenerateKey(c, rand.Reader)

	hashed := []byte("testing")
	r, s, err := ecdsa.Sign(rand.Reader, priv, hashed)
	if err != nil {
		fmt.Printf("%s: error signing: %s", tag, err)
		return
	}

	if !ecdsa.Verify(&priv.PublicKey, hashed, r, s) {
		fmt.Printf("%s: Verify failed", tag)
	}

	hashed[0] ^= 0xff
	if ecdsa.Verify(&priv.PublicKey, hashed, r, s) {
		fmt.Printf("%s: Verify always works!", tag)
	}

	fmt.Printf("done (%s)\n", time.Since(start))
}

func TestSignAndVerify() {
	testSignAndVerify(elliptic.P224(), "p224")
	testSignAndVerify(elliptic.P256(), "p256")
}
