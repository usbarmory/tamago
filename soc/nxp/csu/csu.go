// NXP i.MX Central Security Unit (CSU) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package csu implements a driver for the NXP Central Security Unit (CSU)
// adopting the following reference specifications:
//   - IMX6ULLRM  - i.MX 6ULL Applications Processor Reference Manual          - Rev 1 2017/11
//   - IMX6ULLSRM - i.MX 6ULL Applications Processor Security Reference Manual - Rev 0 2016/09
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/usbarmory/tamago.
package csu

import (
	"errors"

	"github.com/usbarmory/tamago/internal/reg"
)

// CSU registers
const (
	CSU_CSL0 = 0x00
	CSU_SA   = 0x218
)

// CSU represents the Central Security Unit instance.
type CSU struct {
	// Base register
	Base uint32
	// Clock gate register
	CCGR uint32
	// Clock gate
	CG int

	// control registers
	csl0 uint32
	sa   uint32
}

// Init initializes the Central Security Unit (CSU).
func (hw *CSU) Init() {
	if hw.Base == 0 || hw.CCGR == 0 {
		panic("invalid CSU instance")
	}

	hw.csl0 = hw.Base + CSU_CSL0
	hw.sa = hw.Base + CSU_SA

	// enable clock
	reg.SetN(hw.CCGR, hw.CG, 0b11, 0b11)
}

// GetAccess returns the security access (SA) for one of the 16 masters IDs.
// The lock return value indicates whether the SA is locked for changes until
// the next power cycle.
func (hw *CSU) GetAccess(id int) (secure bool, lock bool, err error) {
	if id < SA_MIN || id > SA_MAX {
		return false, false, errors.New("index out of range")
	}

	val := reg.Get(hw.sa, id*2, 0b11)

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
func (hw *CSU) SetAccess(id int, secure bool, lock bool) (err error) {
	if id < SA_MIN || id > SA_MAX {
		return errors.New("index out of range")
	}

	reg.SetTo(hw.sa, id*2, !secure)

	if lock {
		reg.Set(hw.sa, id*2+1)
	}

	return
}
