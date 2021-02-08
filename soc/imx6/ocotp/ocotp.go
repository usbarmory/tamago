// NXP i.MX6 On-Chip OTP Controller (OCOTP_CTRL) driver
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
// Package ocotp implements a driver for the NXP On-Chip OTP Controller
// (OCOTP_CTRL), included in i.MX6 series SoCs to interface with on-chip fuses,
// including write operation.
//
// WARNING: Fusing SoC OTPs is an **irreversible** action that permanently
// fuses values on the device. This means that any errors in the process, or
// lost fused data such as cryptographic key material, might result in a
// **bricked** device.
//
// The use of this package is therefore **at your own risk**.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/f-secure-foundry/tamago.
package ocotp

import (
	"errors"
	"sync"
	"time"

	"github.com/f-secure-foundry/tamago/internal/reg"
	"github.com/f-secure-foundry/tamago/soc/imx6"
)

// OCOTP registers
// (p2388, 37.5 OCOTP Memory Map/Register Definition, IMX6ULLRM).
const (
	OCOTP_BASE = 0x021bc000

	OCOTP_CTRL          = OCOTP_BASE
	CTRL_WRUNLOCK       = 16
	CTRL_RELOAD_SHADOWS = 10
	CTRL_ERROR          = 9
	CTRL_BUSY           = 8
	CTRL_ADDR           = 0

	OCOTP_CTRL_CLR = OCOTP_BASE + 0x0008
	OCOTP_DATA     = OCOTP_BASE + 0x0020

	OCOTP_VERSION = OCOTP_BASE + 0x0090
	VERSION_MAJOR = 24
	VERSION_MINOR = 16
	VERSION_STEP  = 0

	OCOTP_BANK0_WORD0 = OCOTP_BASE + 0x0400
)

// Configuration constants
const (
	// Value required to unlock OCOTP_DATA register
	// (p2393, OCOTP_CTRLn field descriptions, IMX6ULLRM).
	OCOTP_WRUNLOCK_MAGIC = 0x3e77

	// number of words in each bank
	OCOTP_WORDS_PER_BANK = 8

	// shadow registers address map gap between bank 5 and bank 6
	OCOTP_BANK5_GAP = 0x100
)

var mux sync.Mutex

// Timeout for OCOTP controller operations
var Timeout = 10 * time.Millisecond

// Init initializes the OCOTP controller instance.
func Init() (err error) {
	mux.Lock()
	defer mux.Unlock()

	// enable clock
	reg.SetN(imx6.CCM_CCGR2, imx6.CCGR2_CG6, 0b11, 0b11)

	return
}

// Read returns the value in the argument bank and word location.
func Read(bank int, word int) (value uint32, err error) {
	var banks int

	switch imx6.Model() {
	case "i.MX6UL":
		banks = 16
	case "i.MX6ULL":
		banks = 8
	}

	if bank > banks || word > OCOTP_WORDS_PER_BANK {
		return 0, errors.New("invalid argument")
	}

	// Within the shadow register address map the addresses are spaced 0x10
	// apart.
	offset := 0x10 * uint32(OCOTP_WORDS_PER_BANK*bank+word)

	// Account for the gap in shadow registers address map between bank 5
	// and bank 6.
	if bank > 5 {
		offset += OCOTP_BANK5_GAP
	}

	mux.Lock()
	defer mux.Unlock()

	value = reg.Read(OCOTP_BANK0_WORD0 + offset)

	return
}

// Blow fuses a value in the argument bank and word location.
// (p2384, 37.3.1.3 Fuse and Shadow Register Writes, IMX6ULLRM).
//
// WARNING: Fusing SoC OTPs is an **irreversible** action that permanently
// fuses values on the device. This means that any errors in the process, or
// lost fused data such as cryptographic key material, might result in a
// **bricked** device.
//
// The use of this function is therefore **at your own risk**.
func Blow(bank int, word int, value uint32) (err error) {
	mux.Lock()
	defer mux.Unlock()

	if !reg.WaitFor(Timeout, OCOTP_CTRL, CTRL_BUSY, 1, 0) {
		return errors.New("OCOTP controller busy")
	}

	// FIXME: configure OCOTP_TIMING register. Timings depend on
	// IPG_CLK_ROOT frequency. Default values work for default frequency of
	// 66 MHz.

	// clear error bit
	reg.Set(OCOTP_CTRL_CLR, CTRL_ERROR)
	// set OTP write register
	reg.SetN(OCOTP_CTRL, CTRL_ADDR, 0x7f, uint32(OCOTP_WORDS_PER_BANK*bank+word))
	// enable OTP write access
	reg.SetN(OCOTP_CTRL, CTRL_WRUNLOCK, 0xffff, OCOTP_WRUNLOCK_MAGIC)

	// blow the fuse
	reg.Write(OCOTP_DATA, value)

	if err = checkOp(); err != nil {
		return
	}

	// 2385, 37.3.1.4 Write Postamble, IMX6ULLRM
	time.Sleep(2 * time.Microsecond)

	// ensure update of shadow registers
	err = shadowReload()

	return
}

func checkOp() (err error) {
	if !reg.WaitFor(Timeout, OCOTP_CTRL, CTRL_BUSY, 1, 0) {
		return errors.New("operation timeout")
	}

	if reg.Get(OCOTP_CTRL, CTRL_ERROR, 1) != 0 {
		return errors.New("operation error")
	}

	return
}

// shadowReload reloads memory mapped shadow registers from OTP fuse banks
// (p2383, 37.3.1.1 Shadow Register Reload, IMX6ULLRM).
func shadowReload() (err error) {
	if !reg.WaitFor(Timeout, OCOTP_CTRL, CTRL_BUSY, 1, 0) {
		return errors.New("OCOTP controller busy")
	}

	// clear error bit
	reg.Set(OCOTP_CTRL_CLR, CTRL_ERROR)
	// force re-loading of shadow registers
	reg.Set(OCOTP_CTRL, CTRL_RELOAD_SHADOWS)

	err = checkOp()

	return
}
