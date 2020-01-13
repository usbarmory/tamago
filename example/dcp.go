// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build tamago,arm

package main

import (
	"crypto/aes"
	"fmt"
	"strings"

	"github.com/f-secure-foundry/tamago/imx6"
)

const testVector = "\x75\xf9\x02\x2d\x5a\x86\x7a\xd4\x30\x44\x0f\xee\xc6\x61\x1f\x0a"
const zeroVector = "\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"

func testKeyDerivation() (err error) {
	diversifier := []byte{0xde, 0xad, 0xbe, 0xef}
	iv := make([]byte, aes.BlockSize)

	key, err := imx6.DCP.DeriveKey(diversifier, iv)

	if err != nil {
		return
	} else {
		if strings.Compare(string(key), zeroVector) == 0 {
			err = fmt.Errorf("derivedKey all zeros!\n")
			return
		}

		// if the SoC is secure booted we can only print the result
		if imx6.DCP.SNVS() {
			fmt.Printf("imx6_dcp: derived SNVS key %x\n", key)
			return
		}

		if strings.Compare(string(key), testVector) != 0 {
			err = fmt.Errorf("derivedKey:%x != testVector:%x\n", key, testVector)
			return
		} else {
			fmt.Printf("imx6_dcp: derived test key %x\n", key)
		}
	}

	return
}

func TestDCP() {
	imx6.DCP.Init()

	if err := testKeyDerivation(); err != nil {
		fmt.Printf("imx6_dcp: error, %v\n", err)
	}
}
