// i.MX Central Security Unit (CSU) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package csu

import (
	"errors"

	"github.com/usbarmory/tamago/internal/reg"
)

func checkArgs(periph int, slave int) (err error) {
	if periph < CSL_MIN || periph > CSL_MAX {
		return errors.New("peripheral index out of range")
	}

	if slave < 0 || slave > 1 {
		return errors.New("slave index out of range")
	}

	return
}

// GetSecurityLevel returns the config security level (CSL) registers for a
// peripheral slave. The lock return value indicates whether the CSL is locked
// for changes until the next power cycle.
func GetSecurityLevel(periph int, slave int) (csl uint8, lock bool, err error) {
	if err = checkArgs(periph, slave); err != nil {
		return
	}

	val := reg.Read(CSU_CSL0 + uint32(4*periph))
	csl = uint8((val >> (CSL_S2 * slave)) & 0xff)

	if uint8((val>>(CSL_S1_LOCK+CSL_S2*slave))&0b1) == 1 {
		lock = true
	}

	return
}

// SetSecurityLevel sets the config security level (CSL) registers for a
// peripheral slave. The lock argument controls whether the CSL is locked for
// changes until the next power cycle.
func SetSecurityLevel(periph int, slave int, csl uint8, lock bool) (err error) {
	if err = checkArgs(periph, slave); err != nil {
		return
	}

	addr := CSU_CSL0 + uint32(4*periph)

	reg.SetN(addr, CSL_S2*slave, 0xff, uint32(csl))

	if lock {
		reg.Set(addr, CSL_S1_LOCK+CSL_S2*slave)
	}

	return
}
