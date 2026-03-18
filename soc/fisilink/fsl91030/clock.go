// Fisilink FSL91030 clock control
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package fsl91030

// Clock frequencies for the FSL91030/Nuclei UX600.
//
// The timer runs at 32768 Hz (RTCCLK). The high-frequency clocks (200 MHz
// and 100 MHz) and the CPU clock (400 MHz nominal) are set by hardware
// defaults or by the flashboot assembly stub (tools/flashboot.s) before the
// TamaGo runtime starts. There is no runtime PLL reconfiguration interface
// exposed to software.
const (
	RTCCLK   = 32768     // 32768 Hz - Timer/RTC clock frequency
	HFCLK    = 200000000 // 200 MHz - High-frequency clock
	HFCLK2   = 100000000 // 100 MHz - Secondary high-frequency clock
	CPU_FREQ = 400000000 // 400 MHz - CPU frequency (nominal)
)

// Freq returns the RISC-V core frequency.
func Freq() uint32 {
	return CPU_FREQ
}
