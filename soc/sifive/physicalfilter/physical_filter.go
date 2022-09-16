// SiFive Physical Filter (DevicePMP) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package physicalfilter implements a driver for SiFive physical-filter IP
// adopting the following reference specifications:
//   * FU740C00RM - SiFive FU740-C000 Manual - v1p2 2021/03/25
//
// This package is only meant to be used with `GOOS=tamago GOARCH=riscv64` as
// supported by the TamaGo framework for bare metal Go on RISC-V SoCs, see
// https://github.com/usbarmory/tamago.
package physicalfilter

import (
	"errors"
	"sync"

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
)

// DevicePMP constants
// (p194, 23.2.1 Physical Filter Registers, FU740C00RM).
const (
	DEVICE_PMP_0 = 0x00

	PMP_ADDR_HI = 10
	PMP_L       = 63 // lock
	PMP_A       = 59 // address-matching mode
	PMP_W       = 57 // write access
	PMP_R       = 56 // read access

	PMP_A_OFF = false // Null region (disabled)
	PMP_A_TOR = true  // Top of range
)

// PhysicalFilter represents a Physical Filter instance.
type PhysicalFilter struct {
	sync.Mutex

	// Base register
	Base uint32
}

// ReadPMP returns the Device Physical Memory Protection register,
// configuration and address, for the relevant index.
func (hw *PhysicalFilter) ReadPMP(i int) (addr uint64, r bool, w bool, a bool, l bool, err error) {
	if i > 3 {
		err = errors.New("invalid PMP index")
		return
	}

	pmp := reg.Read64(uint64(hw.Base) + uint64(8*i))

	addr = bits.Get64(&pmp, PMP_ADDR_HI, 0x1ffffff) << 4

	r = bits.Get64(&pmp, PMP_R, 1) == 1
	w = bits.Get64(&pmp, PMP_W, 1) == 1
	a = bits.Get64(&pmp, PMP_A, 1) == 1
	l = bits.Get64(&pmp, PMP_L, 1) == 1

	return
}

// WritePMP sets the Device Physical Memory Protection register, configuration
// and address, for the relevant index.
func (hw *PhysicalFilter) WritePMP(i int, addr uint64, r bool, w bool, a bool, l bool) (err error) {
	if i > 3 {
		return errors.New("invalid PMP index")
	}

	hw.Lock()
	defer hw.Unlock()

	pmp := reg.Read64(uint64(hw.Base) + uint64(8*i))

	bits.SetN64(&pmp, PMP_ADDR_HI, 0x1ffffff, addr>>4)

	bits.SetTo64(&pmp, PMP_R, r)
	bits.SetTo64(&pmp, PMP_W, w)
	bits.SetTo64(&pmp, PMP_A, a)
	bits.SetTo64(&pmp, PMP_L, l)

	return
}
