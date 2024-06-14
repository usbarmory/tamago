// NXP i.MX6UL OCRAM support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package imx6ul

import (
	"errors"

	"github.com/usbarmory/tamago/internal/reg"
)

// On-Chip Random-Access Memory
const (
	OCRAM_START = 0x00900000
	OCRAM_SIZE  = 0x20000
)

const (
	GPR10_OCRAM_TZ_ADDR = 11
	GPR10_OCRAM_TZ_EN   = 10
)

// SetOCRAMProtection defines the OCRAM memory region subject to TrustZone
// protection by means of CSU Peripheral Access Policy.
//
// Once set any memory access within the start argument and the end of the
// OCRAM follows the config security level (CSL) set by the CSU (see
// csu.SetSecurityLevel()).
func SetOCRAMProtection(start uint32) (err error) {
	if start < OCRAM_START || start >= OCRAM_START + OCRAM_SIZE {
		return errors.New("address outside OCRAM memory range")
	}

	if start % 4096 != 0 {
		return errors.New("address must be 4K bytes aligned")
	}

	start -= OCRAM_START

	reg.SetN(IOMUXC_GPR_GPR10, GPR10_OCRAM_TZ_ADDR, 0x1f, start / 4096)
	reg.Set(IOMUXC_GPR_GPR10, GPR10_OCRAM_TZ_EN)

	return
}
