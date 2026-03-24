// Fisilink FSL91030 System Clock and Reset Control
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package fsl91030

import "github.com/usbarmory/tamago/internal/reg"

// System Clock Control (SYSCLK, base 0xE084C000) and System Reset Control
// (SYSRST, base 0xE084E000) are FSL91030-specific registers that are not
// documented in the FSL91030M Register Specification PDF. Their addresses and
// bit assignments are derived exclusively from the Linux Ethernet driver
// (xy1000_net.c, lines 63-73) and the U-Boot Ethernet driver (xy1000_eth.c):
//
//	#define SYSCLK_CTRL    0xE084C000
//	#define SYSRST_CTRL    0xE084E000
//	#define CLK_GEN_CSR    0x24        // clock gate register offset
//	#define BLK_SFT_RST_CSR 0x00      // soft reset register offset
//	#define AXI_EMAC_CLK_EN  18       // bit 18: Ethernet MAC clock gate
//	#define SFT_RSTN_AXI_EMAC 18      // bit 18: Ethernet MAC soft reset
//	#define SFT_RSTN_DMA     19       // bit 19: DMA soft reset
//
// RESET POLARITY: The reset bits are active-low ("RSTN" = reset-negated).
//
//	bit = 0 → peripheral held in reset
//	bit = 1 → peripheral released from reset (normal operation)
//
// CLOCK GATE: Active-high.
//
//	bit = 0 → clock gated off
//	bit = 1 → clock enabled
//
// Note: the Linux driver defines these helper functions but the probe path
// never initialises the pointer fields, so the clock/reset helpers are dead
// code in the shipped driver. The Ethernet MAC is expected to be already
// clocked and out of reset when the kernel starts (done by the bootloader).
// This driver provides the correct implementation for bare-metal use.
const (
	// SYSCLK register offsets (from SYSCLK_CTRL base)
	SYSCLK_CLK_CSR = 0x24 // CLK_GEN_CSR: peripheral clock gate register

	// SYSRST register offsets (from SYSRST_CTRL base)
	SYSRST_BLK_CSR = 0x00 // BLK_SFT_RST_CSR: peripheral soft reset register

	// Bit positions (same bit controls both clock gate and reset)
	SYSCTL_ETH_BIT = 18 // Bit 18: Ethernet MAC AXI clock / soft reset
	SYSCTL_DMA_BIT = 19 // Bit 19: DMA soft reset (defined but unused in Linux driver)
)

// SysCtl manages peripheral clock gating and reset sequencing for the
// FSL91030 SoC.
type SysCtl struct {
	// ClockBase is the MMIO base address of the System Clock Control block
	// (0xE084C000).
	ClockBase uint32

	// ResetBase is the MMIO base address of the System Reset Control block
	// (0xE084E000).
	ResetBase uint32
}

// EnableClock enables the clock for the peripheral at bit position pos in the
// clock gate register at ClockBase+SYSCLK_CLK_CSR.
func (s *SysCtl) EnableClock(pos int) {
	reg.Set(s.ClockBase+SYSCLK_CLK_CSR, pos)
}

// DisableClock gates the clock off for the peripheral at bit position pos.
func (s *SysCtl) DisableClock(pos int) {
	reg.Clear(s.ClockBase+SYSCLK_CLK_CSR, pos)
}

// AssertReset holds a peripheral in reset by clearing the active-low reset bit
// at bit position pos in the soft reset register at ResetBase+SYSRST_BLK_CSR.
func (s *SysCtl) AssertReset(pos int) {
	reg.Clear(s.ResetBase+SYSRST_BLK_CSR, pos)
}

// DeassertReset releases a peripheral from reset by setting the active-low
// reset bit at bit position pos.
func (s *SysCtl) DeassertReset(pos int) {
	reg.Set(s.ResetBase+SYSRST_BLK_CSR, pos)
}

// ResetPeripheral performs a full reset cycle: assert reset (clear RSTN bit),
// enable clock (clock must run for reset to propagate), then release reset
// (set RSTN bit). The clock and reset bit positions are given separately as
// they use different registers, though for the Ethernet MAC both happen to be
// bit 18 in their respective registers.
func (s *SysCtl) ResetPeripheral(clkPos, rstPos int) {
	s.AssertReset(rstPos)
	s.EnableClock(clkPos)
	s.DeassertReset(rstPos)
}

// EnableEthernet enables the Ethernet MAC AXI clock and releases the MAC from
// soft reset (bit 18 in both SYSCLK_CLK_CSR and SYSRST_BLK_CSR).
func (s *SysCtl) EnableEthernet() {
	s.ResetPeripheral(SYSCTL_ETH_BIT, SYSCTL_ETH_BIT)
}

// DisableEthernet gates the Ethernet MAC AXI clock.
func (s *SysCtl) DisableEthernet() {
	s.DisableClock(SYSCTL_ETH_BIT)
}
