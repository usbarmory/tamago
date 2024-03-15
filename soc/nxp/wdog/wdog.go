// NXP Watchdog Timer (WDOG) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package bee implements a driver for the NXP Watchdog Timer (WDOG)
// adopting the following reference specifications:
//   - IMX6ULLRM - i.MX 6ULL Applications Processor Reference Manual - Rev 1 2017/11
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/usbarmory/tamago.
package wdog

import (
	"sync"

	"github.com/usbarmory/tamago/internal/reg"
)

// WDOG registers
const (
	WDOGx_WCR = 0x00
	WCR_WT    = 8
	WCR_SRE   = 6
	WCR_WDA   = 5
	WCR_SRS   = 4
	WCR_WDT   = 3
	WCR_WDE   = 2

	WDOGx_WSR  = 0x02

	WDOGx_WRSR = 0x04
	WRSR_POR   = 4
	WRSR_TOUT  = 1
	WRSR_SFTW  = 0

	WDOGx_WICR = 0x06
	WICR_WIE   = 15
	WICR_WTIS  = 14
	WICR_WICT  = 0

	WDOGx_WMCR = 0x08
	WMCR_PDE   = 0
)

// WDOG service sequence
const (
	wsr_seq1 = 0x5555
	wsr_seq2 = 0xaaaa
)

// WDOG represents a Watchdog Timer instance.
type WDOG struct {
	sync.Mutex

	// Module index
	Index int
	// Base register
	Base uint32
	// Clock gate register
	CCGR uint32
	// Clock gate
	CG int
	// Interrupt ID
	IRQ int

	// control registers
	wcr  uint32
	wsr  uint32
	wrsr uint32
	wicr uint32
	wmcr uint32
}

// Init initializes a Watchdog Timer instance. The initialization is required
// within 16 seconds of reset de-assertion to clear the power-down counter
// event.
func (hw *WDOG) Init() {
	hw.Lock()
	defer hw.Unlock()

	if hw.Base == 0 || hw.CCGR == 0 {
		panic("invalid WDOG module instance")
	}

	hw.wcr  = hw.Base + WDOGx_WCR
	hw.wsr  = hw.Base + WDOGx_WSR
	hw.wrsr = hw.Base + WDOGx_WRSR
	hw.wicr = hw.Base + WDOGx_WICR
	hw.wmcr = hw.Base + WDOGx_WMCR

	// enable clock
	reg.SetN(hw.CCGR, hw.CG, 0b11, 0b11)

	// p4085, 59.5.3 Power-down counter event, IMX6ULLRM
	reg.Clear16(hw.wmcr, WMCR_PDE)
}

// EnableInterrupt enables interrupt generation before the Watchdog timeout
// event per argument delay. The delay must be specified in milliseconds with
// 127500 as maximum value, the timeout resolution is 500ms.
func (hw *WDOG) EnableInterrupt(delay int) {
	reg.SetN16(hw.wicr, WICR_WICT, 0xffff, 1 << WICR_WIE | uint16(delay/500))
}

// ClearInterrupt clears the interrupt status register.
func (hw *WDOG) ClearInterrupt() {
	reg.Set16(hw.wicr, WICR_WTIS)
}

// EnableTimeout activates the Watchdog Timer to trigger a reset after the
// argument timeout. The timeout must be specified in milliseconds with 128000
// as maximum value, the timeout resolution is 500ms. The timeout can be
// prevented, or reconfigured, with Service().
func (hw *WDOG) EnableTimeout(timeout int) {
	hw.Lock()
	defer hw.Unlock()

	reg.SetN16(hw.wcr, WCR_WT, 0xff, uint16(timeout/500-1))
	reg.Set16(hw.wcr, WCR_WDT)
	reg.Set16(hw.wcr, WCR_WDE)
}

// Service prevents the timeout condition on a previously enabled Watchdog.
func (hw *WDOG) Service(timeout int) {
	hw.Lock()
	defer hw.Unlock()

	// In case we are a TrustZone Watchdog the Normal World OS might
	// disable our clock, which keeps the timeout but prevents servicing,
	// therefore we re-enable the clock.
	reg.SetN(hw.CCGR, hw.CG, 0b11, 0b11)

	// update timeout
	reg.SetN16(hw.wcr, WCR_WT, 0xff, uint16(timeout/500-1))

	if reg.Get16(hw.wicr, WICR_WIE, 1) == 1 {
		// clear interrupt status
		reg.Set16(hw.wicr, WICR_WTIS)
	}

	// perform service sequence
	reg.Write16(hw.wsr, wsr_seq1)
	reg.Write16(hw.wsr, wsr_seq2)
}

// Reset asserts the watchdog reset signal.
func (hw *WDOG) Reset() {
	reg.Clear16(hw.wcr, WCR_WDA)
}

// SoftwareReset asserts the watchdog software reset signal.
func (hw *WDOG) SoftwareReset() {
	reg.Set16(hw.wcr, WCR_SRE)
	reg.Clear16(hw.wcr, WCR_SRS)
}

// ResetSource reads the watchdog reset status register which records the
// source of the output reset assertion.
func (hw *WDOG) ResetSource() uint16 {
	return reg.Read16(hw.wrsr)
}
