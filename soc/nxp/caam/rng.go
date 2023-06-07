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
	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/internal/rng"
)

func (hw *CAAM) initRNG() {
	// initialize RNG
	op1 := Operation{}
	op1.SetDefaults()
	op1.OpType(OPTYPE_ALG_CLASS1)
	op1.Algorithm(ALG_RNG, 0)
	op1.State(AS_INITIALIZE)

	// enable Prediction Resistance (PR)
	bits.Set(&op1.Word0, 1)

	// wait for previous operation
	jmp := Jump{}
	jmp.SetDefaults()
	jmp.Class(1)
	jmp.Offset(1)

	// clear Class 1 Mode Register
	c1m := Load{}
	c1m.SetDefaults()
	c1m.Destination(CLRW)
	c1m.Immediate(1 << C0CWR_C1M)

	// initialize JDKEK, TDKEK and TDSK
	op2 := Operation{}
	op2.SetDefaults()
	op2.OpType(OPTYPE_ALG_CLASS1)
	op2.Algorithm(ALG_RNG, (1 << AAI_RNG_SK))

	jd := op1.Bytes()
	jd = append(jd, jmp.Bytes()...)
	jd = append(jd, c1m.Bytes()...)
	jd = append(jd, op2.Bytes()...)

	hw.jr.add(nil, jd)
}

// GetRandomData returns len(b) random bytes gathered from the CAAM TRNG.
func (hw *CAAM) GetRandomData(b []byte) {
	hw.Lock()
	defer hw.Unlock()

	// TRNG access through RTENT registers prevents RNG access in CAAM
	// commands, enable only as needed.
	reg.Set(hw.rtmctl, RTMCTL_TRNG_ACC)
	defer reg.Clear(hw.rtmctl, RTMCTL_TRNG_ACC)

	read := 0
	need := len(b)

	for read < need {
		if hw.rtenta == hw.rtent0 {
			for reg.Get(hw.rtmctl, RTMCTL_ENT_VAL, 1) == 0 {
				// wait for valid entropy
			}
		}

		read = rng.Fill(b, read, reg.Read(hw.rtenta))

		if hw.rtenta == hw.rtent15 {
			hw.rtenta = hw.rtent0
		} else {
			hw.rtenta += 4
		}
	}
}
