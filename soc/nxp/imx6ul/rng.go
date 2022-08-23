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
)

//go:linkname initRNG runtime.initRNG
func initRNG() {
	if Native && Family == IMX6ULL {
		RNGB.Init()
		rng.GetRandomDataFn = RNGB.GetRandomData
	} else {
		rng.GetRandomDataFn = rng.GetLCGData
	}
}
