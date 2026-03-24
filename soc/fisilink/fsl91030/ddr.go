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
// Register offsets derived from the vendor's freeloader.S (vega-loader-entire),
// which was used as a reference for register addresses and timing values.
// The initialization sequence is implemented in:
//   - tools/flashboot.s: standalone assembly stub (standard boot path)
//   - boot_riscv64.s: Go assembly cpuinit override (linkcpuinit build tag)
//   - InitDDR() below: Go-callable function (available post-runtime-start)
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

// InitDDR initializes the DDR SDRAM controller (MilkV Vega board SDRAM timing).
//
// In the standard boot flow, flashboot.s runs DDR init before the Go runtime
// starts. This function is available for reconfiguration at runtime; callers
// must call DisableCache before and EnableCache after.
//
//go:nosplit
func InitDDR()
