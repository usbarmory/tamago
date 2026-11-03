// SpacemiT K1 Watchdog Timer driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package k1

import (
	"errors"
	"time"

	"github.com/usbarmory/tamago/soc/spacemit/uart"
)

// The write access sequence and the control values follow the vendor U-Boot
// driver (drivers/watchdog/spacemit_wdt.c).
const (
	WDT_BASE = 0xd4080000

	// Write access sequence
	WDT_WFAR     = 0xb0
	WDT_WFAR_KEY = 0xbaba
	WDT_WSAR     = 0xb4
	WDT_WSAR_KEY = 0xeb10

	WDT_ENABLE  = 0xb8
	WDT_TIMEOUT = 0xbc
	WDT_STATUS  = 0xc0
	WDT_RESET   = 0xc8

	// WDT_ENABLE values
	WDT_RUN  = 0x3
	WDT_HALT = 0x0

	// WDT_RESET value, arms the reset request for the next timeout
	WDT_RESET_ARM    = 0x1
	WDT_STATUS_CLEAR = 0x0

	// Counter frequency, the timeout register counts at 256 Hz as the
	// vendor driver assumes (WDT_CLK_FREQ).
	WDT_FREQ = 256

	// Start control, a register outside the block register file: the
	// vendor device tree gives the watchdog a second entry at
	// 0xd4051020 (MPMU_BASE + 0x1020) whose bit 4 releases the counter.
	WDT_START    = 0xd4051020
	WDT_START_EN = 1 << 4
)

// Main PMU watchdog clock and reset control
//
// Both the clock gate and the reset of the watchdog live in a single MPMU
// register, as the vendor clock driver (ccu-k1x.c) and reset
// driver (reset-spacemit-k1x.c) declare.
const (
	MPMU_WDTPCR = 0x200

	WDT_CLK_EN       = 0x3
	WDT_RESET_ASSERT = 1 << 2
)

// wdtMaxCounter bounds the programmable timeout. The width of the timeout
// register is not documented, so a full 32-bit counter is assumed here, which at
// WDT_FREQ is just under 194 days.
const wdtMaxCounter = 0xffffffff

// WDT represents a Watchdog Timer instance.
type WDT struct {
	Base     uint32
	StartReg uint32
}

func (hw *WDT) write(off uint32, val uint32) {
	uart.Write(hw.Base+WDT_WFAR, WDT_WFAR_KEY)
	uart.Write(hw.Base+WDT_WSAR, WDT_WSAR_KEY)
	uart.Write(hw.Base+off, val)
}

// Init ungates the watchdog clock and deasserts its reset.
//
// It is called by WDT.Start, and is only needed on its own to inspect the
// controller before starting it.
func (hw *WDT) Init() {
	pcr := uint32(MPMU_BASE + MPMU_WDTPCR)

	val := uart.Read(pcr)
	val |= WDT_CLK_EN
	val &^= WDT_RESET_ASSERT

	uart.Write(pcr, val)
}

// Start enables the watchdog, resetting the SoC when the argument timeout
// elapses without an intervening WDT.Feed.
//
// The counter is derived as `timeout * WDT_FREQ`, a timeout shorter than one
// counter tick therefore rounds down and a zero timeout expires immediately,
// which is what WDT.ForceReset relies on.
func (hw *WDT) Start(timeout time.Duration) error {
	if timeout < 0 {
		return errors.New("watchdog timeout can not be negative")
	}
	counter := int64(timeout) * WDT_FREQ / int64(time.Second)

	if counter < 0 || counter > wdtMaxCounter {
		return errors.New("watchdog timeout out of range")
	}

	hw.Init()

	hw.write(WDT_STATUS, WDT_STATUS_CLEAR)
	hw.write(WDT_TIMEOUT, uint32(counter))
	hw.write(WDT_ENABLE, WDT_RUN)
	hw.write(WDT_RESET, WDT_RESET_ARM)

	// The start control is a plain register, it is not covered by the
	// access key sequence.
	uart.Write(hw.StartReg, uart.Read(hw.StartReg)|WDT_START_EN)

	return nil
}

func (hw *WDT) Stop() {
	hw.write(WDT_ENABLE, WDT_HALT)
}

func (hw *WDT) Feed() {
	hw.write(WDT_STATUS, WDT_STATUS_CLEAR)
	hw.write(WDT_RESET, WDT_RESET_ARM)
}

func (hw *WDT) Status() uint32 {
	return uart.Read(hw.Base + WDT_STATUS)
}

func (hw *WDT) Timeout() time.Duration {
	counter := uart.Read(hw.Base + WDT_TIMEOUT)

	return time.Duration(counter) * time.Second / WDT_FREQ
}

// ForceReset resets the SoC through the watchdog and does not return.
//
// Unlike the vendor driver, whose `expire_now` reprograms the controller but
// leaves the start control untouched, the full start sequence is issued so
// that the reset occurs whether or not the counter was already running.
func (hw *WDT) ForceReset() error {
	if err := hw.Start(0); err != nil {
		return err
	}

	select {}
}
