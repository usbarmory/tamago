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
	WCR_WDE   = 2

	WDOGx_WSR = 0x02

	WDOGx_WICR = 0x06
	WICR_WIE   = 15
	WICR_WTIS  = 14

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
	// Interrupt number
	IRQ int

	// control registers
	wcr  uint32
	wsr  uint32
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

	hw.wcr = hw.Base + WDOGx_WCR
	hw.wsr = hw.Base + WDOGx_WSR
	hw.wicr = hw.Base + WDOGx_WICR
	hw.wmcr = hw.Base + WDOGx_WMCR

	// enable clock
	reg.SetN(hw.CCGR, hw.CG, 0b11, 0b11)

	// p4085, 59.5.3 Power-down counter event, IMX6ULLRM
	reg.Clear16(hw.wmcr, WMCR_PDE)
}

// EnableInterrupt enables interrupt generation on timeout events.
func (hw *WDOG) EnableInterrupt() {
	reg.Set16(hw.wicr, WICR_WIE)
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
	reg.Set16(hw.wcr, WCR_WDE)
}

// Service prevents the timeout condition on a previously enabled Watchdog.
func (hw *WDOG) Service(timeout int) {
	hw.Lock()
	defer hw.Unlock()

	if hw.Index == 2 {
		// WDOG2 is the TrustZone Watchdog (TZ WDOG), used to prevent
		// resource starvation by Normal World OS.
		//
		// The Normal World OS might disable its clock, keeping the
		// timeout but preventing servicing, therefore we re-enable the
		// clock.
		reg.SetN(hw.CCGR, hw.CG, 0b11, 0b11)
	}

	// update timeout
	reg.SetN16(hw.wcr, WCR_WT, 0xff, uint16(timeout/500-1))

	// perform service sequence
	reg.Write16(hw.wsr, wsr_seq1)
	reg.Write16(hw.wsr, wsr_seq2)
}

// Reset asserts a system reset signal.
func (hw *WDOG) SoftwareReset() {
	reg.Set16(hw.wcr, WCR_SRE)
	reg.Clear16(hw.wcr, WCR_SRS)
}
