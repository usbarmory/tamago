// Fisilink FSL91030 DDR controller
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package fsl91030

// DDR controller registers (base address: 0x10001000)
//
// Register offsets extracted from vega-loader-entire/freeloader.S
// The boot loader performs a complex initialization sequence to configure
// the DDR SDRAM controller with timing parameters, mode registers, and
// refresh settings.
const (
	DDR_STATUS  = 0x14  // Status register (offset 20)
	DDR_MODE    = 0x200 // DDR mode register (offset 512)
	DDR_CTRL    = 0x204 // DDR control register (offset 516)
	DDR_TIMING0 = 0x208 // Timing parameter 0 (offset 520)
	DDR_TIMING1 = 0x20C // Timing parameter 1 (offset 524)
	DDR_TIMING2 = 0x210 // Timing parameter 2 (offset 528)
	DDR_TIMING3 = 0x214 // Timing parameter 3 (offset 532)
	DDR_TIMING4 = 0x218 // Timing parameter 4 (offset 536)
	DDR_TIMING5 = 0x21C // Timing parameter 5 (offset 540)
	DDR_TIMING6 = 0x220 // Timing parameter 6 (offset 544)
	DDR_TIMING7 = 0x224 // Timing parameter 7 (offset 548)
	DDR_TIMING8 = 0x228 // Timing parameter 8 (offset 552)
	DDR_TIMING9 = 0x22C // Timing parameter 9 (offset 556)
	DDR_TIMINGA = 0x230 // Timing parameter A (offset 560)
	DDR_TIMINGB = 0x234 // Timing parameter B (offset 564)
	DDR_TIMINGC = 0x238 // Timing parameter C (offset 568)
	DDR_TIMINGD = 0x23C // Timing parameter D (offset 572)
	DDR_REG0    = 0x260 // Configuration register 0 (offset 608)
	DDR_REG1    = 0x270 // Configuration register 1 (offset 624)
	DDR_REG2    = 0x274 // Configuration register 2 (offset 628)
	DDR_REG3    = 0x278 // Configuration register 3 (offset 632)
	DDR_REG4    = 0x2A8 // Configuration register 4 (offset 680)
	DDR_REG5    = 0x2AC // Configuration register 5 (offset 684)
	DDR_REG6    = 0x330 // Configuration register 6 (offset 816)
	DDR_REG7    = 0x338 // Configuration register 7 (offset 824)
)

// DDR control register bits
const (
	DDR_CTRL_INIT_DONE = 8 // Bit 8: Initialization complete status
)

// DDR status register bits
const (
	DDR_STATUS_READY = 1 // Bit 1: Controller ready status
)

// InitDDR initializes the DDR SDRAM controller.
//
// This function implements the complete DDR initialization sequence from
// freeloader.S (lines 29-101), including:
//   1. Wait for DDR_STATUS ready bit
//   2. Configure DDR_MODE with SDRAM mode parameters
//   3. Set DDR_TIMING0 through DDR_TIMINGD with timing values
//   4. Configure DDR_REG0 through DDR_REG7 with controller parameters
//   5. Set DDR_CTRL to start initialization
//   6. Wait for DDR_CTRL_INIT_DONE bit (bit 8)
//
// WARNING: The timing parameters are hardware-specific and must match
// the DDR chip datasheet and PCB design. These values are from the
// freeloader.S boot loader and may need adjustment for different hardware
// configurations. Incorrect values may cause memory corruption or boot failure.
//
// Register values written (from freeloader.S):
//   - DDR_MODE (0x200): 0x080002FD
//   - DDR_CTRL (0x204): 0x00000000, then 0x00000001 to trigger init
//   - DDR_TIMING0-D: Various timing parameters
//   - DDR_REG0-7: Configuration registers
//
// Note: If TamaGo is loaded by the boot loader, DDR is already initialized
// and this function is not needed. This is only required if TamaGo runs as
// the first-stage boot loader.
func InitDDR() {
	initDDR()
}

//go:nosplit
func initDDR()
