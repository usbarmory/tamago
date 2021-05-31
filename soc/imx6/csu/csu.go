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
	CSU_SA   = CSU_BASE + 0x218
)

// Init initializes the Central Security Unit (CSU).
func Init() {
	// enable clock
	reg.SetN(imx6.CCM_CCGR1, imx6.CCGR1_CG14, 0b11, 0b11)
}

// GetAccess returns the security access (SA) for one of the 16 masters IDs.
// The lock return value indicates whether the SA is locked for changes until
// the next power cycle.
func GetAccess(id int) (secure bool, lock bool, err error) {
	if id < SA_MIN || id > SA_MAX {
		return false, false, errors.New("index out of range")
	}

	val := reg.Get(CSU_SA, id*2, 0b11)

	if val&0b01 == 0 {
		secure = true
	}

	if val&0b10 != 0 {
		lock = true
	}

	return
}

// SetAccess configures the security access (SA) for one of the 16 masters IDs.
// The lock argument controls whether the SA is locked for changes until the
// next power cycle.
func SetAccess(id int, secure bool, lock bool) (err error) {
	if id < SA_MIN || id > SA_MAX {
		return errors.New("index out of range")
	}

	if secure {
		reg.Clear(CSU_SA, id*2)
	} else {
		reg.Set(CSU_SA, id*2)
	}

	if lock {
		reg.Set(CSU_SA, id*2+1)
	}

	return
}
