// NXP i.MX6UL RNG initialization
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package imx6ul

import (
	"encoding/binary"
	"time"
	_ "unsafe"

	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/internal/rng"
	"github.com/usbarmory/tamago/soc/nxp/caam"
	"github.com/usbarmory/tamago/soc/nxp/rngb"
)

//go:linkname initRNG runtime.initRNG
func initRNG() {
	if !Native {
		drbg := &rng.DRBG{}
		binary.LittleEndian.PutUint64(drbg.Seed[:], uint64(time.Now().UnixNano()))
		rng.GetRandomDataFn = drbg.GetRandomData
		return
	}

	switch Model() {
	case "i.MX6UL":
		// Cryptographic Acceleration and Assurance Module
		CAAM = &caam.CAAM{
			Base:            CAAM_BASE,
			CCGR:            CCM_CCGR0,
			CG:              CCGRx_CG5,
			DeriveKeyMemory: dma.Default(),
		}
		CAAM.Init()

		// The CAAM TRNG is too slow for direct use, therefore
		// we use it to seed an AES-CTR based DRBG.
		drbg := &rng.DRBG{}
		CAAM.GetRandomData(drbg.Seed[:])

		rng.GetRandomDataFn = drbg.GetRandomData
	case "i.MX6ULL", "i.MX6ULZ":
		// True Random Number Generator
		RNGB = &rngb.RNGB{
			Base: RNGB_BASE,
		}
		RNGB.Init()

		rng.GetRandomDataFn = RNGB.GetRandomData
	}
}
