// TrustZone Address Space Controller (TZASC) driver
// https://github.com/f-secure-foundry/tamago
//
// IP: ARM CoreLink™ TrustZone Address Space Controller TZC-380
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package tzasc implements a driver for the TrustZone Address Space Controller
// (TZASC) included in NXP i.MX6ULL/i.MX6ULZ SoCs.
//
// Note that the TZASC must be initialized early in the boot process, see
// TZASC_BYPASS for information.
//
// The driver is based on the following reference specifications:
//   * TZC-380 TRM - CoreLink™ TrustZone Address Space Controller TZC-380 - Revision: r0p1
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/f-secure-foundry/tamago.
package tzasc

import (
	"errors"

	"github.com/f-secure-foundry/tamago/bits"
	"github.com/f-secure-foundry/tamago/internal/reg"
)

// TZASC imx6 specific registers
const (
	IOMUXC_GPR_GPR1       = 0x020e4004
	GPR1_TZASC1_BOOT_LOCK = 23

	// TZASC_BYPASS represents the register that allows to enable the TZASC
	// monitoring of DDR transactions.
	//
	// To use the TZASC the bypass must be disabled early in the boot
	// process, before DDR use.
	//
	// To do so the register can be written in the board DCD file (e.g.
	// imximage.cfg in usbarmory package):
	// 	`DATA 4 0x020e4024 0x00000001`
	//
	// This is a one time operation, until the next power-up cycle.
	TZASC_BYPASS = 0x020e4024

	TZASC_BASE = 0x021d0000
)

// TZASC registers
// (p37, Table 3-1 Register summary, TZC-380 TRM).
const (
	TZASC_CONF   = TZASC_BASE + 0x000
	CONF_REGIONS = 0

	TZASC_ACTION          = TZASC_BASE + 0x004
	TZASC_LOCKDOWN_RANGE  = TZASC_BASE + 0x008
	TZASC_LOCKDOWN_SELECT = TZASC_BASE + 0x00c
	TZASC_SEC_INV_EN      = TZASC_BASE + 0x034

	TZASC_REGION_SETUP_LOW_0  = TZASC_BASE + 0x100
	TZASC_REGION_SETUP_HIGH_0 = TZASC_BASE + 0x104

	TZASC_REGION_ATTRS_0 = TZASC_BASE + 0x108
	REGION_ATTRS_SP      = 28
	REGION_ATTRS_SIZE    = 1
	REGION_ATTRS_EN      = 0

	SIZE_MIN = 0b001110
	SIZE_MAX = 0b111111
)

// TZASC security permissions,
// (p28, Table 2-4, TZC-380 TRM).
const (
	// Secure Read Access bit
	SP_SW_RD = 3
	// Secure Write Access bit
	SP_SW_WR = 2
	// NonSecure Read Access bit
	SP_NW_RD = 1
	// NonSecure Write Access bit
	SP_NW_WR = 0
)

// Regions returns the number of regions that the TZASC provides.
func Regions() int {
	return int(reg.Get(TZASC_CONF, CONF_REGIONS, 0xf)) + 1
}

// EnableSecurityInversion allows configuration of arbitrary security
// permissions, disabling automatic enabling of secure access on non-secure
// only permissions
// (p49, 3.2.12 Security Inversion Enable Register, TZC-380 TRM).
func EnableSecurityInversion() {
	reg.Set(TZASC_SEC_INV_EN, 0)
}

// EnableRegion configures a TZASC region with the argument start address, size
// and security permissions, for region 0 only security permissions are
// relevant.
func EnableRegion(n int, start uint32, size int, sp int) (err error) {
	var attrs uint32
	var s uint32

	if n < 0 || n+1 > Regions() {
		return errors.New("invalid region index")
	}

	if reg.Read(TZASC_BYPASS) != 1 {
		return errors.New("TZASC inactive (bypass detected)")
	}

	if n == 0 {
		reg.SetN(TZASC_REGION_ATTRS_0, REGION_ATTRS_SP, 0b1111, uint32(sp))
		return
	}

	if start%(1<<15) != 0 {
		return errors.New("incompatible start address")
	}

	if start != 0 && (start%uint32(size)) != 0 {
		return errors.New("start address must be a multiple of its region size")
	}

	if sp > 0b1111 {
		return errors.New("invalid security permissions")
	}

	// size = 2^(s+1)
	for i := uint32(SIZE_MIN); i <= SIZE_MAX; i++ {
		if size == (1 << (i + 1)) {
			s = i
			break
		}
	}

	if s == 0 {
		return errors.New("incompatible region size")
	}

	bits.SetN(&attrs, REGION_ATTRS_SP, 0b1111, uint32(sp))
	bits.SetN(&attrs, REGION_ATTRS_SIZE, 0b111111, s)
	bits.Set(&attrs, REGION_ATTRS_EN)

	off := uint32(0x10 * n)

	reg.Write(TZASC_REGION_SETUP_LOW_0+off, start&0xffff8000)
	reg.Write(TZASC_REGION_SETUP_HIGH_0+off, 0)
	reg.Write(TZASC_REGION_ATTRS_0+off, attrs)

	return
}

// DisableRegion disables a TZASC region.
func DisableRegion(n int) (err error) {
	if n < 0 || n+1 > Regions() {
		return errors.New("invalid region index")
	}

	if reg.Read(TZASC_BYPASS) != 1 {
		return errors.New("TZASC inactive (bypass detected)")
	}

	reg.Clear(TZASC_REGION_ATTRS_0+uint32(0x10*n), REGION_ATTRS_EN)

	return
}

// Lock enables TZASC secure boot lock register writing restrictions
// (p30, 2.2.8 Preventing writes to registers and using secure_boot_lock, TZC-380 TRM).
func Lock() {
	reg.Write(TZASC_LOCKDOWN_RANGE, 0xffffffff)
	reg.Write(TZASC_LOCKDOWN_SELECT, 0xffffffff)
	reg.Set(GPR1_TZASC1_BOOT_LOCK, 23)
}
