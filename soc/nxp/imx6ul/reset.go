// NXP i.MX6UL system reset controller (SRC) support
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

	SRC_GPR10                    = SRC_SCR + 0x44
	GPR10_PERSIST_SECONDARY_BOOT = 30
)

// PersistSecondaryBoot controls whether the primary (false) or secondary boot
// image (true) should be selected after a software reset.
func PersistSecondaryBoot(enable bool) {
	// tested to work correctly only with SetWarmReset(false)
	reg.SetTo(SRC_GPR10, GPR10_PERSIST_SECONDARY_BOOT, enable)
}

// SetWarmReset controls whether warm reset sources are enabled (true) or if
// they should generate a cold reset (false).
func SetWarmReset(enable bool) {
	reg.SetTo(SRC_SCR, SCR_WARM_RESET_ENABLE, enable)
}

// Reset asserts the global watchdog reset causing the SoC to restart with a
// cold reset.
//
// Note that only the SoC itself is guaranteed to restart as, depending on the
// board hardware layout, the system might remain powered (which might not be
// desirable). See respective board packages for cold reset options.
func Reset() {
	SetWarmReset(false)

	// assert software reset
	WDOG1.SoftwareReset()
}
