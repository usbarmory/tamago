// NXP i.MX6UL ARM debug signals
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package imx6ul

import (
	"github.com/usbarmory/tamago/internal/reg"
)

const (
	IOMUXC_GPR_GPR10 = 0x020e4028
	GPR10_DBG_CLK_EN = 1
	GPR10_DBG_EN     = 0
)

// Debug controls ARM invasive and non-invasive debug functionalities.
func Debug(enable bool) {
	reg.SetTo(IOMUXC_GPR_GPR10, GPR10_DBG_CLK_EN, enable)
	reg.SetTo(IOMUXC_GPR_GPR10, GPR10_DBG_EN, enable)
}
