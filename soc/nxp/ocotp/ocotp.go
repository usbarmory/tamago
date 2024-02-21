// NXP i.MX6 On-Chip OTP Controller (OCOTP_CTRL) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package ocotp implements a driver for the NXP On-Chip OTP Controller
// (OCOTP_CTRL), which provides an interface to on-chip fuses for read/write
// operation, adopting the following reference specifications:
//   - IMX6ULLRM - i.MX 6ULL Applications Processor Reference Manual - Rev 1 2017/11
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
// https://github.com/usbarmory/tamago.
package ocotp

import (
	"errors"
	"sync"
	"time"

	"github.com/usbarmory/tamago/internal/reg"
)

// OCOTP registers
// (p2388, 37.5 OCOTP Memory Map/Register Definition, IMX6ULLRM).
const (
	OCOTP_CTRL          = 0x0000
	CTRL_WRUNLOCK       = 16
	CTRL_RELOAD_SHADOWS = 10
	CTRL_ERROR          = 9
	CTRL_BUSY           = 8
	CTRL_ADDR           = 0

	OCOTP_CTRL_CLR = 0x0008
	OCOTP_DATA     = 0x0020
)

// Configuration constants
const (
	// WordSize represents the number of bytes per OTP word.
	WordSize = 4
	// BankSize represents the number of words per OTP bank.
	BankSize = 8
	// Timeout is the default timeout for OCOTP operations.
	Timeout = 10 * time.Millisecond
)

type OCOTP struct {
	sync.Mutex

	// Base register
	Base uint32
	// Bank base register (bank 0, word 0)
	BankBase uint32
	// Banks size
	Banks int
	// Clock gate register
	CCGR uint32
	// Clock gate
	CG int
	// Timeout for OCOTP controller operations
	Timeout time.Duration

	// control registers
	ctrl     uint32
	ctrl_clr uint32
	data     uint32
}

// Init initializes the OCOTP controller instance.
func (hw *OCOTP) Init() {
	hw.Lock()
	defer hw.Unlock()

	if hw.Base == 0 || hw.BankBase == 0 || hw.CCGR == 0 {
		panic("invalid OCOTP instance")
	}

	if hw.Timeout == 0 {
		hw.Timeout = Timeout
	}

	hw.ctrl = hw.Base + OCOTP_CTRL
	hw.ctrl_clr = hw.Base + OCOTP_CTRL_CLR
	hw.data = hw.Base + OCOTP_DATA

	// enable clock
	reg.SetN(hw.CCGR, hw.CG, 0b11, 0b11)
}

// Read returns the value in the argument bank and word location.
func (hw *OCOTP) Read(bank int, word int) (value uint32, err error) {
	if bank > hw.Banks || word > BankSize {
		return 0, errors.New("invalid argument")
	}

	// Within the shadow register address map the addresses are spaced 0x10
	// apart.
	offset := 0x10 * uint32(BankSize*bank+word)

	// Account for the gap in shadow registers address map between bank 5
	// and bank 6.
	if bank > 5 {
		offset += 0x100
	}

	hw.Lock()
	defer hw.Unlock()

	value = reg.Read(hw.BankBase + offset)

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
func (hw *OCOTP) Blow(bank int, word int, value uint32) (err error) {
	hw.Lock()
	defer hw.Unlock()

	if !reg.WaitFor(hw.Timeout, hw.ctrl, CTRL_BUSY, 1, 0) {
		return errors.New("OCOTP controller busy")
	}

	// We do not configure the OCOTP_TIMING register. Timings depend on
	// IPG_CLK_ROOT frequency. Default values work for default frequency of
	// 66 MHz.

	// p2393, OCOTP_CTRLn field descriptions, IMX6ULLRM

	// clear error bit
	reg.Set(hw.ctrl_clr, CTRL_ERROR)
	// set OTP write register
	reg.SetN(hw.ctrl, CTRL_ADDR, 0x7f, uint32(BankSize*bank+word))
	// enable OTP write access
	reg.SetN(hw.ctrl, CTRL_WRUNLOCK, 0xffff, 0x3e77)

	// blow the fuse
	reg.Write(hw.data, value)

	if err = hw.checkOp(); err != nil {
		return
	}

	// 2385, 37.3.1.4 Write Postamble, IMX6ULLRM
	time.Sleep(2 * time.Microsecond)

	// ensure update of shadow registers
	return hw.shadowReload()
}

func (hw *OCOTP) checkOp() (err error) {
	if !reg.WaitFor(hw.Timeout, hw.ctrl, CTRL_BUSY, 1, 0) {
		return errors.New("operation timeout")
	}

	if reg.Get(hw.ctrl, CTRL_ERROR, 1) != 0 {
		return errors.New("operation error")
	}

	return
}

// shadowReload reloads memory mapped shadow registers from OTP fuse banks
// (p2383, 37.3.1.1 Shadow Register Reload, IMX6ULLRM).
func (hw *OCOTP) shadowReload() (err error) {
	if !reg.WaitFor(hw.Timeout, hw.ctrl, CTRL_BUSY, 1, 0) {
		return errors.New("OCOTP controller busy")
	}

	// clear error bit
	reg.Set(hw.ctrl_clr, CTRL_ERROR)
	// force re-loading of shadow registers
	reg.Set(hw.ctrl, CTRL_RELOAD_SHADOWS)

	return hw.checkOp()
}
