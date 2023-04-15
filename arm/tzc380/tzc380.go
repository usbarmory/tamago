// ARM TrustZone Address Space Controller TZC-380 driver
// https://github.com/usbarmory/tamago
//
// IP: ARM CoreLink™ TrustZone Address Space Controller TZC-380
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package tzc380 implements a driver for the ARM TrustZone Address Space
// Controller TZC-380.
//
// Note that the TZASC must be initialized early in the boot process, see
// TZASC.Bypass for information.
//
// The driver is based on the following reference specifications:
//   - TZC-380 TRM - CoreLink™ TrustZone Address Space Controller TZC-380 - Revision: r0p1
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/usbarmory/tamago.
package tzc380

import (
	"errors"

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
)

// TZASC registers
// (p37, Table 3-1 Register summary, TZC-380 TRM).
const (
	TZASC_CONF   = 0x000
	CONF_REGIONS = 0

	TZASC_LOCKDOWN_RANGE  = 0x008
	TZASC_LOCKDOWN_SELECT = 0x00c
	TZASC_SEC_INV_EN      = 0x034

	TZASC_REGION_SETUP_LOW_0  = 0x100
	TZASC_REGION_SETUP_HIGH_0 = 0x104

	TZASC_REGION_ATTRS_0 = 0x108
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

// TZASC represents the TrustZone Address Space Controller instance.
type TZASC struct {
	// Base register
	Base uint32

	// The bypass register controls the TZASC monitoring of DDR
	// transactions.
	//
	// To use the TZASC, the bypass must be disabled early in the boot
	// process, before DDR use. This is a one time operation, until the
	// next power-up cycle.
	//
	// To do so the register can be written in the board DCD file (e.g.
	// imximage.cfg in usbarmory package):
	// 	`DATA 4 0x020e4024 0x00000001`
	//
	// The register must be specified in the TZASC instance for
	// verification purposes.
	Bypass uint32

	// Secure Boot Lock signal (see 2.2.8 TZC-380 TRM)
	SecureBootLockReg uint32
	SecureBootLockPos int

	// control registers
	conf                uint32
	lockdown_range      uint32
	lockdown_select     uint32
	sec_inv_en          uint32
	region_setup_low_0  uint32
	region_setup_high_0 uint32
	region_attrs_0      uint32
}

// Init initializes the TrustZone Address Space Controller (TZASC).
func (hw *TZASC) Init() {
	if hw.Base == 0 || hw.Bypass == 0 || hw.SecureBootLockReg == 0 {
		panic("invalid TZASC instance")
	}

	hw.conf = hw.Base + TZASC_CONF
	hw.lockdown_range = hw.Base + TZASC_LOCKDOWN_RANGE
	hw.lockdown_select = hw.Base + TZASC_LOCKDOWN_SELECT
	hw.sec_inv_en = hw.Base + TZASC_SEC_INV_EN
	hw.region_setup_low_0 = hw.Base + TZASC_REGION_SETUP_LOW_0
	hw.region_setup_high_0 = hw.Base + TZASC_REGION_SETUP_HIGH_0
	hw.region_attrs_0 = hw.Base + TZASC_REGION_ATTRS_0
}

// Regions returns the number of regions that the TZASC provides.
func (hw *TZASC) Regions() int {
	return int(reg.Get(hw.conf, CONF_REGIONS, 0xf)) + 1
}

// EnableSecurityInversion allows configuration of arbitrary security
// permissions, disabling automatic enabling of secure access on non-secure
// only permissions
// (p49, 3.2.12 Security Inversion Enable Register, TZC-380 TRM).
func (hw *TZASC) EnableSecurityInversion() {
	reg.Set(hw.sec_inv_en, 0)
}

// EnableRegion configures a TZASC region with the argument start address, size
// and security permissions, for region 0 only security permissions are
// relevant.
func (hw *TZASC) EnableRegion(n int, start uint32, size uint32, sp int) (err error) {
	var attrs uint32
	var s uint32

	if n < 0 || n+1 > hw.Regions() {
		return errors.New("invalid region index")
	}

	if reg.Read(hw.Bypass) != 1 {
		return errors.New("TZASC inactive (bypass detected)")
	}

	if n == 0 {
		reg.SetN(hw.region_attrs_0, REGION_ATTRS_SP, 0b1111, uint32(sp))
		return
	}

	if start%(1<<15) != 0 {
		return errors.New("incompatible start address")
	}

	if start != 0 && (start%size) != 0 {
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

	reg.Write(hw.region_setup_low_0+off, start&0xffff8000)
	reg.Write(hw.region_setup_high_0+off, 0)
	reg.Write(hw.region_attrs_0+off, attrs)

	return
}

// DisableRegion disables a TZASC region.
func (hw *TZASC) DisableRegion(n int) (err error) {
	if n < 0 || n+1 > hw.Regions() {
		return errors.New("invalid region index")
	}

	if reg.Read(hw.Bypass) != 1 {
		return errors.New("TZASC inactive (bypass detected)")
	}

	reg.Clear(hw.region_attrs_0+uint32(0x10*n), REGION_ATTRS_EN)

	return
}

// Lock enables TZASC secure boot lock register writing restrictions
// (p30, 2.2.8 Preventing writes to registers and using secure_boot_lock, TZC-380 TRM).
func (hw *TZASC) Lock() {
	reg.Write(hw.lockdown_range, 0xffffffff)
	reg.Write(hw.lockdown_select, 0xffffffff)
	reg.Set(hw.SecureBootLockReg, hw.SecureBootLockPos)
}
