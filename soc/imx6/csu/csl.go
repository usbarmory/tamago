// i.MX Central Security Unit (CSU) driver
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package csu

import (
	"errors"

	"github.com/f-secure-foundry/tamago/internal/reg"
)

func checkArgs(peripheral int, slave int) (err error) {
	if peripheral < CSL_MIN || peripheral > CSL_MAX {
		return errors.New("peripheral index out of range")
	}

	if slave < 0 || slave > 1 {
		return errors.New("slave index out of range")
	}

	return
}

// GetSecurityLevel returns the config security level (CSL) registers for a
// peripheral slave.
func GetSecurityLevel(peripheral int, slave int) (csl uint8, err error) {
	if err = checkArgs(peripheral, slave); err != nil {
		return
	}

	val := reg.Read(CSU_CSL0 + uint32(4*peripheral))
	shift := CSL_S2 * slave

	return uint8((val >> shift) & 0xff), nil
}

// SetSecurityLevel sets the config security level (CSL) registers for a
// peripheral slave. The lock argument controls whether the CSL is locked for
// changes until the next power cycle.
func SetSecurityLevel(peripheral int, slave int, csl uint8, lock bool) (err error) {
	if err = checkArgs(peripheral, slave); err != nil {
		return
	}

	addr := CSU_CSL0 + uint32(4*peripheral)

	reg.SetN(addr, CSL_S2*slave, 0xff, uint32(csl))

	if lock {
		reg.Set(addr, CSL_S1_LOCK+CSL_S2*slave)
	}

	return
}
