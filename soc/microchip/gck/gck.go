// Microchip LAN969x Generic Clock (GCK) configuration
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package gck implements configuration of the Microchip LAN969x Generic Clock
// (GCK_CFG) registers that supply peripheral functional clocks.
//
// The following specification is adopted:
//   - Microchip - LAN9694/LAN9696/LAN9698 Datasheet - DS00005048E (02-27-25)
//
// This package is only meant to be used with `GOOS=tamago` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package gck

import (
	"errors"

	"github.com/usbarmory/tamago/internal/reg"
)

// GCK_CFG register fields
const (
	GCK_ENA            = 0
	GCK_SRC_SEL        = 8
	GCK_SRC_SEL_MASK   = 0x3
	GCK_PRESCALER      = 16
	GCK_PRESCALER_MASK = 0xff
)

// Prescaler returns the smallest prescaler whose output frequency, the
// parentHz source divided by prescaler + 1, does not exceed targetHz.
func Prescaler(parentHz uint32, targetHz uint32) (prescaler uint32, err error) {
	if parentHz == 0 || targetHz == 0 {
		return 0, errors.New("invalid clock frequency")
	}

	divider := (uint64(parentHz) + uint64(targetHz) - 1) / uint64(targetHz)

	if divider > GCK_PRESCALER_MASK+1 {
		return 0, errors.New("clock prescaler out of range")
	}

	prescaler = uint32(divider - 1)

	return
}

// Enable disables the generic clock at the GCK_CFG register address, selects
// source 0 and the given prescaler, then enables it.
func Enable(addr uint32, prescaler uint32) {
	reg.Clear(addr, GCK_ENA)
	reg.SetN(addr, GCK_SRC_SEL, GCK_SRC_SEL_MASK, 0)
	reg.SetN(addr, GCK_PRESCALER, GCK_PRESCALER_MASK, prescaler)
	reg.Set(addr, GCK_ENA)
}
