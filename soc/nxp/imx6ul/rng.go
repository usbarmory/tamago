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
	_ "unsafe"

	"github.com/usbarmory/tamago/internal/rng"
	"github.com/usbarmory/tamago/soc/nxp/caam"
	"github.com/usbarmory/tamago/soc/nxp/rngb"
)

//go:linkname initRNG runtime.initRNG
func initRNG() {
	if !Native {
		rng.GetRandomDataFn = rng.GetLCGData
		return
	}

	switch Model() {
	case "i.MX6UL":
		// Cryptographic Acceleration and Assurance Module (UL only)
		CAAM = &caam.CAAM{
			Base: CAAM_BASE,
			CCGR: CCM_CCGR0,
			CG:   CCGRx_CG5,
		}

		CAAM.Init()

		rng.GetRandomDataFn = CAAM.GetRandomData
	case "i.MX6ULL", "i.MX6ULZ":
		// True Random Number Generator (ULL/ULZ only)
		RNGB = &rngb.RNGB{
			Base: RNGB_BASE,
		}

		RNGB.Init()

		rng.GetRandomDataFn = RNGB.GetRandomData
	}
}
