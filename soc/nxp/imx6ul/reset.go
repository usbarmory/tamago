// NXP i.MX6UL watchdog support
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

// System Reset Controller registers
const (
	SRC_SCR               = 0x020d8000
	SCR_WARM_RESET_ENABLE = 0
)

// Reset asserts the global watchdog reset causing the SoC to restart.
//
// Note that only the SoC itself is guaranteed to restart as, depending on the
// board hardware layout, the system might remain powered (which might not be
// desirable). See respective board packages for cold reset options.
func Reset() {
	// disable warm reset
	reg.Clear(SRC_SCR, SCR_WARM_RESET_ENABLE)

	// assert software reset
	WDOG1.SoftwareReset()
}
