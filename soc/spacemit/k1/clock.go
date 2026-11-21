// SpacemiT K1 clock control
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package k1

// Oscillator frequencies
//
// The K1 clock tree is derived from a 32K RTC clock and a 24M OSC clock, fed
// to three PLLs which generate all CPU and peripheral frequencies (p72,
// Section 2.11 Clock & Reset, K1 Datasheet).
const (
	// RTCCLK is the always-on 32.768 kHz RTC oscillator.
	RTCCLK = 32768

	// OSCCLK is the 24 MHz VCXO, which also drives the RISC-V time CSR
	// (the X60 core implements Zicntr, p9, Section 2.1.2.2 Features, K1
	// Datasheet) used as time source by this package, see Counter. It
	// matches the timebase-frequency device tree property set by the
	// vendor boot stages.
	OSCCLK = 24000000

	// UARTCLK is the UART baud rate generator input clock as programmed by
	// the vendor boot stages, yielding a divisor of 8 for the 115200 8N1
	// default console configuration.
	// The value is derived from the divisor latch left by the boot ROM
	// loaded first stage, read back with uart.UART.Divisor: a divisor of
	// 8 for 115200 baud gives 8 * 16 * 115200, the standard 14.7456 MHz
	// serial clock.
	UARTCLK = 14745600
)
