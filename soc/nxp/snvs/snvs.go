// NXP Secure Non-Volatile Storage (SNVS) support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package snvs implements helpers for NXP Secure Non-Volatile Storage (SNVS)
// configuration adopting the the following reference specifications:
//   * IMX6ULLRM  - i.MX 6ULL Applications Processor Reference Manual          - Rev 1 2017/11
//   * IMX6ULLSRM - i.MX 6ULL Applications Processor Security Reference Manual - Rev 0 2016/09
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/usbarmory/tamago.
package snvs

import (
	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
)

// SNVS registers
const (
	SNVS_HPSR           = 0x14
	HPSR_OTPMK_ZERO     = 27
	HPSR_OTPMK_SYNDROME = 16

	HPSR_SSM_STATE    = 8
	SSM_STATE_TRUSTED = 0b1101
	SSM_STATE_SECURE  = 0b1111
)

// SNVS represents the SNVS instance.
type SNVS struct {
	// Base register
	Base uint32
	// Clock gate register
	CCGR uint32
	// Clock gate
	CG int
}

// Init initializes the SNVS controller.
func (hw *SNVS) Init() {
	if hw.Base == 0 || hw.CCGR == 0 {
		panic("invalid SNVS instance")
	}

	// enable clock
	reg.SetN(hw.CCGR, hw.CG, 0b11, 0b11)
}

// Available verifies whether the Secure Non Volatile Storage (SNVS) is
// correctly programmed and in Trusted or Secure state (indicating that Secure
// Boot is enabled).
//
// The unique OTPMK internal key is available only when Secure Boot (HAB) is
// enabled, otherwise a Non-volatile Test Key (NVTK), identical for each SoC,
// is used.
func (hw *SNVS) Available() bool {
	if hw.Base == 0 {
		return false
	}

	hpsr := reg.Read(hw.Base + SNVS_HPSR)

	// ensure that the OTPMK has been correctly programmed
	if bits.Get(&hpsr, HPSR_OTPMK_ZERO, 1) != 0 || bits.Get(&hpsr, HPSR_OTPMK_SYNDROME, 0x1ff) != 0 {
		return false
	}

	switch bits.Get(&hpsr, HPSR_SSM_STATE, 0b1111) {
	case SSM_STATE_TRUSTED, SSM_STATE_SECURE:
		return true
	default:
		return false
	}
}
