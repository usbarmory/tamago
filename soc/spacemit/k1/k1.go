// SpacemiT K1 configuration and support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package k1 provides support to Go bare metal unikernels, written using the
// TamaGo framework, on the SpacemiT Key Stone K1 System-on-Chip (SoC).
// This package is only meant to be used with `GOOS=tamago GOARCH=riscv64` as
// supported by the TamaGo framework for bare metal Go on RISC-V SoCs, see
// https://github.com/usbarmory/tamago.
package k1

import (
	"github.com/usbarmory/tamago/riscv64"
	"github.com/usbarmory/tamago/soc/sifive/clint"
	"github.com/usbarmory/tamago/soc/spacemit/uart"
)

// Peripheral registers
//
// The K1 Datasheet documents the SoC memory map only for the Multi-Function
// Pin Registers (p128, Section 4.7 Multi-Function Pin Register (MFPRs)), the
// remaining base addresses match the vendor SDK (OpenSBI, U-Boot and Linux
// device tree) definitions.
const (
	// DRAM window, backed by the LPDDR4/LPDDR4x/LPDDR3 controller
	DRAM_BASE = 0x00000000

	// Main PMU, clock and reset control
	MPMU_BASE = 0xd4050000

	// APB peripheral clock and reset control
	APBC_BASE = 0xd4015000

	// Application Processor PMU, AXI/AHB clock and reset control
	APMU_BASE = 0xd4282800

	// Multi-Function Pin Registers, 129 registers with a 0x4 stride
	// (p128, Section 4.7, K1 Datasheet)
	MFPR_BASE = 0xd401e000

	// General Purpose I/O, four banks of 32 pins
	// (p71, Section 2.9.7 GPIO, K1 Datasheet)
	GPIO_BASE = 0xd4019000

	// Serial ports, the ten 16550 compatible controllers
	// (p62, Section 2.7.7 UART Interface, K1 Datasheet) are mapped
	// contiguously with an UART_STRIDE spacing.
	UART0_BASE  = 0xd4017000
	UART_STRIDE = 0x100

	// Core-Local Interruptor
	CLINT_BASE = 0xe4000000

	// Platform-Level Interrupt Controller, 256 external interrupt sources
	// (p11, Section 2.1.3 Interrupt Controller, K1 Datasheet)
	PLIC_BASE = 0xe0000000
	PLIC_SIZE = 0x4000000
)

// Peripheral instances
var (
	// RISC-V core (SpacemiT X60, RV64GCVB)
	RV64 = &riscv64.CPU{
		Counter: Counter,
		// required before Init()
		TimerMultiplier: 1,
		TimerOffset:     1,
	}

	// Core-Local Interruptor, used for inter-processor interrupts only.
	//
	// The block mtime register does not advance on this SoC and is
	// therefore unsuitable as time source, Counter reads the `time` CSR
	// instead. RTCCLK is set for consistency with the CSR input clock, it
	// only affects the (unused) clint.CLINT.Nanotime.
	CLINT = &clint.CLINT{
		Base:   CLINT_BASE,
		RTCCLK: OSCCLK,
	}

	// UART0, the K1 debug console
	//
	// Clock is deliberately left unset, so that uart.UART.Init keeps the
	// divisor latch programmed by the boot ROM loaded first stage rather
	// than recomputing it: the controller input clock is not documented by
	// the K1 Datasheet, and reprogramming the baud rate generator from a
	// wrong value silences the console. See uart.UART.Divisor to recover the clock
	// from the inherited configuration.
	UART0 = &uart.UART{
		Index:    0,
		Base:     UART0_BASE,
		Baudrate: uart.UART_DEFAULT_BAUDRATE,
	}

	Watchdog = &WDT{
		Base:     WDT_BASE,
		StartReg: WDT_START,
	}
)

func uartClock() uint32 {
	return UARTCLK
}

func Model() string {
	return "K1"
}
