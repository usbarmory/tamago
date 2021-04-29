// NXP i.MX6UL/i.MX6ULL/i.MX6Q support
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

// Watchdog control registers, 32-bit access should be avoided as all registers
// are 16-bit.
const (
	WDOG1_WCR = 0x020bc000
	WDOG2_WCR = 0x020c0000
	WDOG3_WCR = 0x021e4000

	WCR_SRE = 6
	WCR_SRS = 4
)

// System Reset Controller registers
const (
	SRC_SCR               = 0x020d8000
	SCR_WARM_RESET_ENABLE = 0
)

// Reset asserts the global watchdog reset causing the SoC to restart (warm
// reset).
//
// Note that only the SoC itself restarts, while the board remains powered
// (which might not be desirable). See respective board packages for cold reset
// options.
func Reset() {
	// enable warm reset
	reg.Clear(SRC_SCR, SCR_WARM_RESET_ENABLE)

	// enable software reset extension
	reg.Set16(WDOG1_WCR, WCR_SRE)

	// assert system reset signal
	reg.Clear16(WDOG1_WCR, WCR_SRS)
}
