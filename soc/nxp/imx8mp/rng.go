// NXP i.MX8MP configuration and support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package imx8mp

import (
	"encoding/binary"
	"time"
	_ "unsafe"

	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/internal/rng"
	"github.com/usbarmory/tamago/soc/nxp/caam"
)

//go:linkname initRNG runtime/goos.InitRNG
func initRNG() {
	_, Family, _ = SiliconVersion()

	// only emulated targets have been tested so far
	Native = false

	if !Native {
		drbg := &rng.DRBG{}
		binary.LittleEndian.PutUint64(drbg.Seed[:], uint64(time.Now().UnixNano()))
		rng.GetRandomDataFn = drbg.GetRandomData
		return
	}

	switch Family {
	case IMX8MPD, IMX8MPQ:
		// Cryptographic Acceleration and Assurance Module
		CAAM = &caam.CAAM{
			Base:            CAAM_BASE,
			DeriveKeyMemory: dma.Default(),
		}
		CAAM.Init()

		// The CAAM TRNG is too slow for direct use, therefore
		// we use it to seed an AES-CTR based DRBG.
		drbg := &rng.DRBG{}
		CAAM.GetRandomData(drbg.Seed[:])

		rng.GetRandomDataFn = drbg.GetRandomData
	}
}
