// NXP Cryptographic Acceleration and Assurance Module (CAAM) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package caam

import (
	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/internal/rng"
)

// GetRandomData returns len(b) random bytes gathered from the RNGB module.
func (hw *CAAM) GetRandomData(b []byte) {
	hw.Lock()
	defer hw.Unlock()

	read := 0
	need := len(b)

	for read < need {
		if reg.Get(hw.rtmctl, RTMCTL_ENT_VAL, 1) == 1 {
			read = rng.Fill(b, read, reg.Read(hw.rtenta))
		}

		hw.rtenta += 4

		if hw.rtenta == hw.rtent15 {
			hw.rtenta = hw.rtent0
		}
	}
}
