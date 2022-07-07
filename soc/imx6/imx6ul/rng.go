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
	"github.com/usbarmory/tamago/soc/imx6"
)

//go:linkname initRNG runtime.initRNG
func initRNG() {
	if imx6.Family == imx6.IMX6ULL && imx6.Native {
		RNGB.Init()
		rng.GetRandomDataFn = RNGB.GetRandomData
	} else {
		rng.GetRandomDataFn = rng.GetLCGData
	}
}


