// Fisilink FSL91030 clock control
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package fsl91030

// Clock frequencies
//
// The FSL91030/Nuclei UX600 clock configuration:
//   - RTCCLK: 32768 Hz - Timer frequency (from OpenSBI platform.c)
//   - hfclk: 200 MHz - High-frequency clock (from device tree)
//   - hfclk2: 100 MHz - Secondary high-frequency clock (from device tree)
//   - CPU: 400 MHz (nominal, measured dynamically by OpenSBI)
//
// Note: The device tree shows timebase-frequency as 48828 Hz, but OpenSBI
// platform code uses 32768 Hz (UX600_TIMER_FREQ). Using OpenSBI value as
// it's the tested implementation.
//
// Unlike SiFive FU540, there's no visible PRCI (Power, Reset, Clock,
// Interrupt) controller in the device tree for runtime PLL configuration.
// The clocks appear to be configured by the boot loader (freeloader.S)
// or firmware before TamaGo starts.
const (
	RTCCLK   = 32768      // 32768 Hz - Timer/RTC clock frequency
	HFCLK    = 200000000  // 200 MHz - High-frequency clock
	HFCLK2   = 100000000  // 100 MHz - Secondary high-frequency clock
	CPU_FREQ = 400000000  // 400 MHz - CPU frequency (nominal)
)

// Freq returns the RISC-V core frequency.
func Freq() uint32 {
	// CPU frequency is fixed at 400 MHz
	// TODO: If runtime PLL configuration is needed, implement PRCI equivalent
	return CPU_FREQ
}
