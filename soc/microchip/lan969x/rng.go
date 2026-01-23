// Microchip LAN969x configuration and support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package lan969x

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/internal/rng"
)

//go:linkname initRNG runtime.initRNG
func initRNG() {
	TRNG.Init()
	rng.GetRandomDataFn = TRNG.GetRandomData
}
