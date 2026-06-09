// Fisilink FSL91030 configuration and support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package fsl91030 provides support to Go bare metal unikernels written using
// the TamaGo framework.
//
// The package implements initialization and drivers for the Fisilink FSL91030
// System-on-Chip (SoC), a RV64IMAFDC processor based on the Nuclei UX600 core
// with sv39 MMU, running at 400 MHz. Its peripherals are largely
// SiFive-compatible (UART, GPIO, SPI).
//
// This package is only meant to be used with `GOOS=tamago GOARCH=riscv64` as
// supported by the TamaGo framework for bare metal Go on RISC-V SoCs, see
// https://github.com/usbarmory/tamago.
package fsl91030

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/riscv64"
	"github.com/usbarmory/tamago/soc/sifive/clint"
	"github.com/usbarmory/tamago/soc/sifive/uart"
)

// Peripheral registers
const (
	// Nuclei timer; the CLINT-compatible mtime/mtimecmp interface sits at
	// offset 0x1000 (mtime at 0x200cff8).
	NUCLEI_TIMER_BASE = 0x2000000
	CLINT_BASE        = NUCLEI_TIMER_BASE + 0x1000

	// software reset key written to NUCLEI_TIMER_BASE+NUCLEI_TIMER_MSFTRST
	NUCLEI_TIMER_MSFTRST     = 0xff0
	NUCLEI_TIMER_MSFTRST_KEY = 0x80000a5f

	// Platform-Level Interrupt Controller
	PLIC_BASE = 0x8000000

	// DDR controller
	DDR_BASE = 0x10019000

	// GPIO block; the Nuclei UX600 variant places the IOF (I/O Function)
	// registers at offsets 0x44 (enable) and 0x48 (select).
	GPIO_BASE    = 0x10011000
	GPIO_IOF_EN  = GPIO_BASE + 0x44
	GPIO_IOF_SEL = GPIO_BASE + 0x48

	// UART0 (USB-C console) and UART1 (rear connector)
	UART0_BASE = 0x10013000
	UART1_BASE = 0x10012000

	// QSPI0 NOR flash controller and its XIP window
	QSPI0_BASE   = 0x10014000
	NOR_XIP_BASE = 0x20000000
	NOR_SIZE     = 64 * 1024 * 1024

	// DRAM
	DRAM_BASE = 0x41000000
	DRAM_SIZE = 240 * 1024 * 1024

	// Watchdog (Andes ATCWDT200)
	WDT_BASE = 0x68000000
)

// Peripheral instances
var (
	// RISC-V core (RV64IMAFDC)
	RV64 = &riscv64.CPU{
		Counter: Counter,
		// required before Init()
		TimerMultiplier: 1,
		TimerOffset:     1,
	}

	// Core-Local Interruptor (CLINT-compatible timer)
	CLINT = &clint.CLINT{
		Base:   CLINT_BASE,
		RTCCLK: RTCCLK,
	}

	// UART0 (USB-C console)
	UART0 = &uart.UART{
		Index:    0,
		Base:     UART0_BASE,
		Clock:    uartClock,
		Baudrate: uart.UART_DEFAULT_BAUDRATE,
		Setup:    uart.UART_SETUP_8N1,
	}

	// UART1 (rear connector)
	UART1 = &uart.UART{
		Index:    1,
		Base:     UART1_BASE,
		Clock:    uartClock,
		Baudrate: uart.UART_DEFAULT_BAUDRATE,
		Setup:    uart.UART_SETUP_8N1,
	}

	// Watchdog timer (Andes ATCWDT200)
	Watchdog = &WDT{Base: WDT_BASE}
)

// UART0 signals are routed to GPIO pin 16 (TX) and pin 17 (RX) via IOF0
// connected to the onboard FTDI channel 0.
const (
	GPIO_UART0_TX = 16
	GPIO_UART0_RX = 17
)

// uartClock returns the UART baud-generator input clock frequency.
func uartClock() uint32 {
	return UARTCLK
}

// InitUARTPinmux routes the UART0 signals to their physical pins by selecting
// IOF0 (clearing the function select bit) and enabling the IOF override. Must
// be called during board initialization before UART0 can be used.
func InitUARTPinmux() {
	reg.Clear(GPIO_IOF_SEL, GPIO_UART0_TX)
	reg.Clear(GPIO_IOF_SEL, GPIO_UART0_RX)
	reg.Set(GPIO_IOF_EN, GPIO_UART0_TX)
	reg.Set(GPIO_IOF_EN, GPIO_UART0_RX)
}

// Model returns the SoC model name.
func Model() string {
	return "FSL91030"
}

// Reset resets the SoC by writing the software-reset key to the Nuclei timer
// MSFTRST register (matching the vendor OpenSBI implementation).
func Reset() {
	reg.Write(NUCLEI_TIMER_BASE+NUCLEI_TIMER_MSFTRST, NUCLEI_TIMER_MSFTRST_KEY)

	// wait for reset
	select {}
}
