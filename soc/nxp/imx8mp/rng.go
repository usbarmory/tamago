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

	"github.com/usbarmory/tamago/internal/rng"
)

//go:linkname initRNG runtime.initRNG
func initRNG() {
	Native = false

	if !Native {
		drbg := &rng.DRBG{}
		binary.LittleEndian.PutUint64(drbg.Seed[:], uint64(time.Now().UnixNano()))
		rng.GetRandomDataFn = drbg.GetRandomData
		return
	}
}
