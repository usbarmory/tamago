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

// Watchdog control registers, 32-bit access should be avoided as all registers
// are 16-bit.
const (
	WDOG1_WCR  = 0x020bc000
	WDOG1_WMCR = 0x020bc008

	WDOG2_WCR  = 0x020c0000
	WDOG2_WMCR = 0x020c0008

	WDOG3_WCR  = 0x021e4000
	WDOG3_WMCR = 0x021e4008

	WCR_SRE  = 6
	WCR_WDA  = 5
	WCR_SRS  = 4
	WMCR_PDE = 0
)

// System Reset Controller registers
const (
	SRC_SCR               = 0x020d8000
	SCR_WARM_RESET_ENABLE = 0
)

func clearWDOG() {
	// Clear the 16 seconds power-down counter event for all watchdogs
	// (p4085, 59.5.3 Power-down counter event, IMX6ULLRM).
	reg.Clear16(WDOG1_WMCR, WMCR_PDE)
	reg.Clear16(WDOG2_WMCR, WMCR_PDE)
	reg.Clear16(WDOG3_WMCR, WMCR_PDE)
}

// Reset asserts the global watchdog reset causing the SoC to restart (warm
// reset).
//
// Note that only the SoC itself is guaranteed to restart as, depending on the
// board hardware layout, the system might remain powered (which might not be
// desirable). See respective board packages for cold reset options.
func Reset() {
	// enable warm reset
	reg.Clear(SRC_SCR, SCR_WARM_RESET_ENABLE)

	// enable software reset extension
	reg.Set16(WDOG1_WCR, WCR_SRE)

	// assert system reset signal
	reg.Clear16(WDOG1_WCR, WCR_SRS)
}
