// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package imx6

import (
	"github.com/f-secure-foundry/tamago/internal/reg"
)

const (
	IOMUXC_GPR_GPR10 = 0x020e4028
	GPR10_DBG_CLK_EN = 1
	GPR10_DBG_EN     = 0
)

// EnableDebug enables the ARM invasive and non-invasive debug functionality.
func EnableDebug() {
	reg.Set(IOMUXC_GPR_GPR10, GPR10_DBG_CLK_EN)
	reg.Set(IOMUXC_GPR_GPR10, GPR10_DBG_EN)
}

// DisableDebug disables the ARM invasive and non-invasive debug functionality.
func DisableDebug() {
	reg.Clear(IOMUXC_GPR_GPR10, GPR10_DBG_CLK_EN)
	reg.Clear(IOMUXC_GPR_GPR10, GPR10_DBG_EN)
}
