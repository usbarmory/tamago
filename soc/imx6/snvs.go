// NXP Secure Non Volatile Storage (SNVS)
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package imx6

import (
	"github.com/f-secure-foundry/tamago/bits"
	"github.com/f-secure-foundry/tamago/internal/reg"
)

const (
	SNVS_HPSR_REG       = 0x020cc014
	HPSR_OTPMK_ZERO     = 27
	HPSR_OTPMK_SYNDROME = 16

	HPSR_SSM_STATE    = 8
	SSM_STATE_TRUSTED = 0b1101
	SSM_STATE_SECURE  = 0b1111
)

// SNVS verifies whether the Secure Non Volatile Storage (SNVS) is available in
// Trusted or Secure state (indicating that Secure Boot is enabled).
//
// The unique OTPMK internal key is available only when Secure Boot (HAB) is
// enabled, otherwise a Non-volatile Test Key (NVTK), identical for each SoC,
// is used.
func SNVS() bool {
	hpsr := reg.Read(SNVS_HPSR_REG)

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
