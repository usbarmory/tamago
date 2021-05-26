// i.MX Central Security Unit (CSU) driver
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package csu implements a driver for the Central Security Unit (CSU) included
// in NXP i.MX6ULL/i.MX6ULZ SoCs.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/f-secure-foundry/tamago.
package csu

import (
	"errors"

	"github.com/f-secure-foundry/tamago/internal/reg"
	"github.com/f-secure-foundry/tamago/soc/imx6"
)

// CSU registers
const (
	CSU_BASE = 0x021c0000

	CSU_CSL0 = CSU_BASE
	CSL_MIN  = 0
	CSL_MAX  = 39

	CSU_SA = CSU_BASE + 0x218
)

// Init initializes the Central Security Unit (CSU).
func Init() {
	// enable clock
	reg.SetN(imx6.CCM_CCGR1, imx6.CCGR1_CG14, 0b11, 0b11)
}

// SetMasterPrivilege configures the access policy for one of the 16 bus
// masters IDs. The lock argument controls whether the CSL is locked for
// changes until the next power cycle.
func SetMasterPrivilege(master int, secure bool, lock bool) (err error) {
	if master < 0 || master > 15 {
		return errors.New("index out of range")
	}

	if secure {
		reg.Clear(CSU_SA, master*2)
	} else {
		reg.Set(CSU_SA, master*2)
	}

	if lock {
		reg.Set(CSU_SA, master*2+1)
	}

	return
}
