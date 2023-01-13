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
	WDOG1_WSR  = 0x020bc002
	WDOG1_WICR = 0x020bc006
	WDOG1_WMCR = 0x020bc008

	WDOG2_WCR  = 0x020c0000
	WDOG2_WSR  = 0x020c0002
	WDOG2_WICR = 0x020c0006
	WDOG2_WMCR = 0x020c0008

	WDOG3_WCR  = 0x021e4000
	WDOG3_WSR  = 0x021e0002
	WDOG3_WICR = 0x021e4006
	WDOG3_WMCR = 0x021e4008

	WCR_WT  = 8
	WCR_SRE = 6
	WCR_WDA = 5
	WCR_SRS = 4
	WCR_WDE = 2

	WICR_WIE  = 16
	WICR_WTIS = 14

	WMCR_PDE = 0
)

// Watchdog interrupts
const (
	WDOG1_IRQ = 32 + 80
	WDOG2_IRQ = 32 + 81
	WDOG3_IRQ = 32 + 11

	TZ_WDOG     = 2
	TZ_WDOG_IRQ = WDOG2_IRQ
)

// System Reset Controller registers
const (
	SRC_SCR               = 0x020d8000
	SCR_WARM_RESET_ENABLE = 0
)

func disablePowerDownCounters() {
	// Clear the 16 seconds power-down counter event for all watchdogs
	// (p4085, 59.5.3 Power-down counter event, IMX6ULLRM).
	reg.Clear16(WDOG1_WMCR, WMCR_PDE)
	reg.Clear16(WDOG2_WMCR, WMCR_PDE)
	reg.Clear16(WDOG3_WMCR, WMCR_PDE)
}

// EnableWatchdog activates a Watchdog to trigger a reset after the argument
// timeout. The timeout must be specified in milliseconds with 128000 as
// maximum value, the timeout resolution is 500ms.
//
// The interrupt enabling is write once, therefore disabling it on subsequent
// calls has no effect.
//
// Calling the function on a previously enabled watchdog performs its service
// sequence to reset its timeout to a new value.
func EnableWatchdog(index int, timeout int, irq bool) {
	var base uint32

	switch index {
	case 1:
		base = WDOG1_WCR
	case 2:
		base = WDOG2_WCR
	case 3:
		base = WDOG3_WCR
	default:
		return
	}

	reg.SetN16(base, WCR_WT, 0xff, uint16(timeout / 500 - 1))

	if reg.Get16(base, WCR_WDE, 1) == 1 {
		reg.Write16(base+2, 0x5555)
		reg.Write16(base+2, 0xaaaa)
		reg.Set16(base+6, WICR_WTIS)
	} else {
		reg.SetTo16(base+6, WICR_WIE, irq)
		reg.Set16(base, WCR_WDE)
	}
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
