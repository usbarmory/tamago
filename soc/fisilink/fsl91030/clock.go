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
// These are fixed hardware clocks; the vendor OpenSBI does not configure any
// PLL (its frequency-measurement code is disabled and it simply assumes the
// values below). The high-frequency clocks and the CPU clock are therefore
// available from reset, with no runtime PLL reconfiguration interface.
const (
	// RTCCLK is the CLINT mtime tick rate, per the vendor device tree
	// (nuclei-ux608.dts: timebase-frequency = 48828, i.e. HFCLK/4096).
	RTCCLK   = 48828     // ~48.828 kHz - Timer/RTC clock frequency
	HFCLK    = 200000000 // 200 MHz - High-frequency clock
	HFCLK2   = 100000000 // 100 MHz - Secondary high-frequency clock
	CPU_FREQ = 400000000 // 400 MHz - CPU frequency (nominal)

	// UARTCLK is the UART baud-generator input clock. Confirmed against the
	// vendor U-Boot, which programs UART0 (0x10013000) DIV = 0x364 (868) for
	// a 115200 console: f_in = 115200 * (868 + 1) ~= 100 MHz (HFCLK2). The
	// SiFive UART formula is baud = f_in / (div + 1).
	UARTCLK = HFCLK2
)

// Freq returns the RISC-V core frequency.
func Freq() uint32 {
	return CPU_FREQ
}
