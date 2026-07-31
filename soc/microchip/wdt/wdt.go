// Synopsys DesignWare APB Watchdog Timer
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package wdt implements a driver for Synopsys DesignWare APB Watchdog Timers
// as integrated in Microchip LAN969x SoCs adopting the following
// specifications:
//   - Microchip - LAN9694/LAN9696/LAN9698 Datasheet - DS00005048E (02-27-25)
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm64` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package wdt

import (
	"sync"

	"github.com/usbarmory/tamago/internal/reg"
)

const (
	WDT_CR    = 0x00
	CR_RMOD   = 1
	CR_WDT_EN = 0

	WDT_TORR = 0x04
	TORR_TOP = 0

	WDT_CCVR = 0x08
	WDT_CRR  = 0x0c

	WDT_STAT = 0x10
	STAT_WDT = 0

	WDT_EOI = 0x14

	WDT_COMP_PARAM_1 = 0xf4
	WDT_COMP_VERSION = 0xf8
	WDT_COMP_TYPE    = 0xfc

	TOP_MIN = 0
	TOP_MAX = 15

	CRR_KICK = 0x76

	// CLOCK_HZ matches the 250 MHz LAN969x fabric clock.
	// p501, 3.47.3 CLOCKING AND RESET, Microchip DS00005048E
	CLOCK_HZ = 250000000
)

// WDT represents a Watchdog Timer instance.
type WDT struct {
	sync.Mutex

	// Base register
	Base uint32
}

// EnableReset starts the watchdog in direct-reset mode. The timeout range
// represents 2^(16+top) watchdog clock cycles.
func (hw *WDT) EnableReset(top int) {
	hw.enable(top, false)
}

// EnableInterrupt starts the watchdog in interrupt-first mode, the timeout
// range represents 2^(16+top) watchdog clock cycles.
//
// The first timeout raises an interrupt and the second timeout resets the
// system.
func (hw *WDT) EnableInterrupt(top int) {
	hw.enable(top, true)
}

func (hw *WDT) enable(top int, interrupt bool) {
	if top < TOP_MIN || top > TOP_MAX {
		panic("invalid watchdog timeout range")
	}

	hw.Lock()
	defer hw.Unlock()

	reg.Write(hw.Base+WDT_TORR, uint32(top))
	reg.SetTo(hw.Base+WDT_CR, CR_RMOD, interrupt)
	reg.Set(hw.Base+WDT_CR, CR_WDT_EN)

	hw.Kick()
}

// Kick restarts the watchdog counter.
func (hw *WDT) Kick() {
	reg.Write(hw.Base+WDT_CRR, CRR_KICK)
}

// Enabled reports whether the watchdog is running.
func (hw *WDT) Enabled() bool {
	return reg.Get(hw.Base+WDT_CR, CR_WDT_EN)
}

// Counter returns the current watchdog counter value.
func (hw *WDT) Counter() uint32 {
	return reg.Read(hw.Base + WDT_CCVR)
}

// Top returns the configured timeout range.
func (hw *WDT) Top() int {
	return int(reg.GetN(hw.Base+WDT_TORR, TORR_TOP, 0xf))
}

// InterruptMode reports whether interrupt-first mode is selected.
func (hw *WDT) InterruptMode() bool {
	return reg.Get(hw.Base+WDT_CR, CR_RMOD)
}

// Status reports whether the watchdog interrupt is active.
func (hw *WDT) Status() bool {
	return reg.Get(hw.Base+WDT_STAT, STAT_WDT)
}

// ClearInterrupt clears the watchdog interrupt without restarting the counter.
func (hw *WDT) ClearInterrupt() {
	reg.Read(hw.Base + WDT_EOI)
}

// Version returns the component version register.
func (hw *WDT) Version() uint32 {
	return reg.Read(hw.Base + WDT_COMP_VERSION)
}

// ComponentType returns the component type register.
func (hw *WDT) ComponentType() uint32 {
	return reg.Read(hw.Base + WDT_COMP_TYPE)
}

// Parameters returns the component parameters register.
func (hw *WDT) Parameters() uint32 {
	return reg.Read(hw.Base + WDT_COMP_PARAM_1)
}
