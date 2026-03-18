// Microchip One Time Programmable Controller (OTPC) support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package otpc implements helpers for the Microchip OTP Controller (OTPC).
//
// WARNING: Fusing SoC OTPs is an **irreversible** action that permanently
// fuses values on the device. This means that any errors in the process, or
// lost fused data such as cryptographic key material, might result in a
// **bricked** device.
//
// The use of this package is therefore **at your own risk**.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm64` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package otpc

import (
	"errors"
	"sync"
	"time"

	"github.com/usbarmory/tamago/internal/reg"
)

// OTPC registers
const (
	OTP_PWR_DN = 0x00
	PWR_DN_N   = 0

	OTP_ADDR_HI   = 0x04
	OTP_ADDR_LO   = 0x08
	OTP_PRGM_DATA = 0x10

	OTP_PRGM_MODE = 0x14
	MODE_BYTE     = 0

	OTP_RD_DATA = 0x18

	OTP_FUNC_CMD = 0x20
	CMD_PROGRAM  = 1
	CMD_READ     = 0

	OTP_CMD_GO = 0x28

	OTP_PASS_FAIL         = 0x2c
	FAIL_READ_PROHIBITED  = 3
	FAIL_WRITE_PROHIBITED = 2
	FAIL_FAIL             = 0

	OTP_STATUS     = 0x30
	STATUS_CPUMPEN = 1
	STATUS_BUSY    = 0
)

// Timeout is the default timeout for OTP operations.
const Timeout = 100 * time.Millisecond

// OTP controller instance
type OTPC struct {
	sync.Mutex

	// Base register
	Base uint32
	// Timeout for OTP controller operations
	Timeout time.Duration
	// OTP size
	Size int
}

func (hw *OTPC) power(up bool) {
	reg.SetTo(hw.Base+OTP_PWR_DN, PWR_DN_N, !up)

	if up {
		reg.Wait(hw.Base+OTP_STATUS, STATUS_CPUMPEN, 1, 0)
	}
}

func (hw *OTPC) command(addr uint32, cmd int, access int) (err error) {
	reg.Write(hw.Base+OTP_ADDR_HI, addr>>8)
	reg.Write(hw.Base+OTP_ADDR_LO, addr&0xff)

	reg.Set(hw.Base+OTP_FUNC_CMD, cmd)
	reg.Write(hw.Base+OTP_CMD_GO, 1)

	timeout := hw.Timeout

	if timeout == 0 {
		timeout = Timeout
	}

	if !reg.WaitFor(timeout, hw.Base+OTP_CMD_GO, 0, 1, 0) {
		return errors.New("command timeout")
	}

	if !reg.WaitFor(timeout, hw.Base+OTP_STATUS, STATUS_BUSY, 1, 0) {
		return errors.New("busy timeout")
	}

	if reg.Get(hw.Base+OTP_PASS_FAIL, access) {
		return errors.New("command prohibited")
	}

	if reg.Get(hw.Base+OTP_PASS_FAIL, FAIL_FAIL) {
		return errors.New("command failure")
	}

	return
}

func (hw *OTPC) read(addr uint32) (b byte, err error) {
	if err = hw.command(addr, CMD_READ, FAIL_READ_PROHIBITED); err != nil {
		return
	}

	return byte(reg.Read(hw.Base + OTP_RD_DATA)), nil
}

func (hw *OTPC) write(addr uint32, b byte) (err error) {
	reg.Set(hw.Base+OTP_PRGM_MODE, MODE_BYTE)
	reg.Write(hw.Base+OTP_PRGM_DATA, uint32(b))

	return hw.command(addr, CMD_PROGRAM, FAIL_WRITE_PROHIBITED)
}

// Read reads a sequence of bytes from OTP memory.
func (hw *OTPC) Read(off int, b []byte) (err error) {
	hw.Lock()
	defer hw.Unlock()

	if off + len(b) >= hw.Size {
		return errors.New("address out of range")
	}

	hw.power(true)
	defer hw.power(false)

	for i := range b {
		if b[i], err = hw.read(uint32(off + i)); err != nil {
			return
		}
	}

	return
}

// Blow writes a sequence of bytes to OTP memory.
//
// WARNING: Fusing SoC OTPs is an **irreversible** action that permanently
// fuses values on the device. This means that any errors in the process, or
// lost fused data such as cryptographic key material, might result in a
// **bricked** device.
//
// The use of this function is therefore **at your own risk**.
func (hw *OTPC) Blow(off int, b []byte) (err error) {
	hw.Lock()
	defer hw.Unlock()

	if off + len(b) >= hw.Size {
		return errors.New("address out of range")
	}

	hw.power(true)
	defer hw.power(false)

	for i := range b {
		if err = hw.write(uint32(off + i), b[i]); err != nil {
			return
		}
	}

	return
}
