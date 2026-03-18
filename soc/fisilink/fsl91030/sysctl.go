// Fisilink FSL91030 System Clock and Reset Control
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package fsl91030

import "github.com/usbarmory/tamago/internal/reg"

// System Clock Control register layout (base address: 0xE084C000).
// System Reset Control register layout (base address: 0xE084E000).
//
// These registers are FSL-specific and are not part of the SiFive or Nuclei
// standard peripheral set. Register bit assignments are derived from the
// FSL91030M Register Specification (FSL91030M寄存器说明书-D) and from
// analysis of the Linux xy1000_net.c Ethernet driver probe sequence.
//
// SYSCLK registers control peripheral clock gating. Writing 1 to a bit
// enables the clock for that peripheral; writing 0 gates it.
//
// SYSRST registers control peripheral reset. Writing 1 to a bit asserts
// (holds in reset) the peripheral; writing 0 releases reset.
//
// NOTE: Bit assignments below are tentative. They match the Ethernet MAC
// reset/clock enable sequence observed in xy1000_net.c. Additional bits
// for UART, SPI, I2C, WDT will be confirmed from the register spec PDF
// and hardware testing.
const (
	// SYSCLK register offsets
	SYSCLK_ETH = 0x00 // Ethernet MAC clock enable register

	// SYSRST register offsets
	SYSRST_ETH = 0x00 // Ethernet MAC reset register

	// Bit positions within SYSCLK_ETH and SYSRST_ETH
	SYSCTL_ETH_BIT = 0 // Bit 0: Ethernet MAC clock / reset
)

// SysCtl manages peripheral clock gating and reset sequencing for the
// FSL91030 SoC.
type SysCtl struct {
	// ClockBase is the base address of the System Clock Control block
	// (0xE084C000).
	ClockBase uint32

	// ResetBase is the base address of the System Reset Control block
	// (0xE084E000).
	ResetBase uint32
}

// EnableClock enables the clock for the peripheral at bit position pos in
// the clock control register at ClockBase+offset.
func (s *SysCtl) EnableClock(offset uint32, pos int) {
	reg.Set(s.ClockBase+offset, pos)
}

// DisableClock gates the clock for the peripheral at bit position pos.
func (s *SysCtl) DisableClock(offset uint32, pos int) {
	reg.Clear(s.ClockBase+offset, pos)
}

// AssertReset holds a peripheral in reset (reset active = bit set).
func (s *SysCtl) AssertReset(offset uint32, pos int) {
	reg.Set(s.ResetBase+offset, pos)
}

// DeassertReset releases a peripheral from reset (reset inactive = bit clear).
func (s *SysCtl) DeassertReset(offset uint32, pos int) {
	reg.Clear(s.ResetBase+offset, pos)
}

// ResetPeripheral performs a full reset cycle for the peripheral at the given
// clock and reset register offsets and bit positions: assert reset, enable
// clock (required to propagate the reset), then deassert reset.
//
// This sequence matches the Ethernet MAC initialization pattern observed in
// the Linux xy1000_net.c driver probe function.
func (s *SysCtl) ResetPeripheral(clkOffset, rstOffset uint32, clkPos, rstPos int) {
	s.AssertReset(rstOffset, rstPos)
	s.EnableClock(clkOffset, clkPos)
	s.DeassertReset(rstOffset, rstPos)
}

// EnableEthernet enables the Ethernet MAC clock and releases it from reset.
func (s *SysCtl) EnableEthernet() {
	s.ResetPeripheral(SYSCLK_ETH, SYSRST_ETH, SYSCTL_ETH_BIT, SYSCTL_ETH_BIT)
}

// DisableEthernet gates the Ethernet MAC clock.
func (s *SysCtl) DisableEthernet() {
	s.DisableClock(SYSCLK_ETH, SYSCTL_ETH_BIT)
}
