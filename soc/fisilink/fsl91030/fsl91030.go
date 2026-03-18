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
// System-on-Chip (SoC), a RV64IMAFDC processor based on the Nuclei UX600
// core with sv39 MMU, running at 400 MHz. Many peripherals are
// SiFive-compatible (UART, GPIO, SPI), while others (Ethernet MAC, Watchdog)
// are FSL-specific.
//
// ISA Note: The FSL91030 implements RV64IMAFDC (64-bit with atomic, single and
// double-precision FPU, and compressed instructions). The standard TamaGo
// RISCV64 compiler (tamago-go) is compatible with this ISA. If running on a
// constrained variant without F/D extensions, use the softfloat branch
// (GOSOFT=1 with tamago1.26.0-73608-softfloat).
//
// This package is only meant to be used with `GOOS=tamago GOARCH=riscv64` as
// supported by the TamaGo framework for bare metal Go on RISC-V SoCs, see
// https://github.com/usbarmory/tamago.
package fsl91030

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/riscv64"
	"github.com/usbarmory/tamago/soc/sifive/clint"
	"github.com/usbarmory/tamago/soc/sifive/gpio"
	"github.com/usbarmory/tamago/soc/sifive/uart"
)

// Peripheral registers
const (
	// Nuclei Timer (base for CLINT-compatible timer)
	// The Nuclei UX600 timer exposes a raw mtime at 0x2000000. The
	// CLINT-compatible mtime/mtimecmp interface sits at +0x1000 (0x2001000),
	// placing mtime at 0x200CFF8 (Base+0xBFF8 as read by the SiFive CLINT driver).
	NUCLEI_TIMER_BASE = 0x2000000
	CLINT_BASE        = NUCLEI_TIMER_BASE + 0x1000 // 0x2001000

	// Platform-Level Interrupt Controller (PLIC)
	// 53 interrupt sources, 7 priority levels
	PLIC_BASE = 0x8000000

	// DDR Controller
	DDR_BASE = 0x10001000

	// GPIO (SiFive-compatible, used for UART/SPI pinmux)
	// Address confirmed by nuclei_ux600fd.dts (gpio@10011000) and OpenSBI.
	// NOTE: 0x10012000 is UART1, not GPIO.
	GPIO_BASE = 0x10011000

	// Serial ports (SiFive UART0 compatible)
	// UART0: console, IRQ 2; UART1: secondary, IRQ 3
	UART0_BASE = 0x10013000
	UART1_BASE = 0x10023000

	// QSPI Flash Controller (SiFive SPI0 compatible)
	// Infineon S25HL512T, 64 MB NOR flash (512Mbit, 3.0V)
	// XIP window: 0x20000000–0x23FFFFFF (64 MB)
	// JEDEC: Manufacturer=0x34, DevID[0]=0x2A (HL-T), DevID[1]=0x1A (512Mb)
	// 4-byte addressing required for offsets >16MB (enter via 0xB7 command).
	// Flash layout:
	//   0x00000000  flashboot stub (64 KB)
	//   0x00010000  TamaGo binary (~60 MB max)
	//   0x03C00000  Config partition (2 MB, log-structured KV)
	//   0x03E00000  OTA staging area (1 MB)
	QSPI0_BASE   = 0x10014000       // SPI controller registers
	NOR_XIP_BASE = 0x20000000       // NOR flash XIP window
	NOR_SIZE     = 64 * 1024 * 1024 // 64 MB (S25HL512T)
	NOR_CFG_OFF  = 0x03C00000       // Config partition offset (XIP-relative)
	NOR_OTA_OFF  = 0x03E00000       // OTA staging area offset (XIP-relative)

	// QSPI1: NAND Flash/MMC (SiFive SPI0 compatible), IRQ 36
	QSPI1_BASE = 0x10016000

	// I2C Controller (OpenCores I2C compatible)
	I2C0_BASE = 0x10018000

	// DRAM
	DRAM_BASE = 0x41000000
	DRAM_SIZE = 240 * 1024 * 1024 // 240 MB (0xF000000)

	// Local Bus Base (from ethernet driver)
	LOCAL_BUS_BASE = 0x60000000

	// Ethernet MAC (FSL-specific, xy1000_eth driver)
	// MAC registers at base + 0x400
	// DMA registers at base + 0x0
	// Interrupts: 10 (RX_END), 11 (RX_REQ), 12 (TX_END)
	ETH_MAC_BASE = 0x67800000

	// Watchdog (FSL-specific)
	WDT_BASE = 0x68000000

	// System Clock Control
	SYSCLK_CTRL_BASE = 0xE084C000

	// System Reset Control
	SYSRST_CTRL_BASE = 0xE084E000
)

// Peripheral instances
var (
	// RISC-V core (RV64IMAFDC)
	RV64 = &riscv64.CPU{
		Counter:         Counter,
		TimerMultiplier: 1,
		// required before Init()
		TimerOffset: 1,
	}

	// Core-Local Interruptor (CLINT-compatible timer)
	// The Nuclei UX600 timer is CLINT-compatible at offset 0x1000
	CLINT = &clint.CLINT{
		Base:   CLINT_BASE,
		RTCCLK: RTCCLK,
	}

	// Serial port 0 (Console)
	UART0 = &uart.UART{
		Index: 0,
		Base:  UART0_BASE,
	}

	// Serial port 1 (Secondary)
	UART1 = &uart.UART{
		Index: 1,
		Base:  UART1_BASE,
	}

	// GPIO0 (SiFive-compatible at 0x10011000, Nuclei UX600 IOF offsets)
	GPIO = &gpio.GPIO{
		Base:         GPIO_BASE,
		IOFENOffset:  gpio.GPIO_IOF_EN_NUCLEI,
		IOFSELOffset: gpio.GPIO_IOF_SEL_NUCLEI,
	}

	// Watchdog timer (Andes ATCWDT200 at 0x68000000, IRQ 8)
	Watchdog = &WDT{Base: WDT_BASE}

	// System Clock and Reset Control
	SystemControl = &SysCtl{
		ClockBase: SYSCLK_CTRL_BASE,
		ResetBase: SYSRST_CTRL_BASE,
	}

	// TODO: QSPI0/QSPI1 support (SiFive SPI0 compatible)
	// QSPI0 (0x10014000): NOR Flash controller (Infineon S25HL512T, 64 MB).
	// QSPI1 (0x10016000): NAND Flash/MMC; requires GPIO IOF configuration.

	// TODO: I2C0 support (OpenCores I2C compatible at 0x10018000)

	// TODO: Ethernet MAC SoC-level instance (xy1000_eth at 0x67800000).
	// The driver is implemented in vega-baremetal/pkg/hal/eth/.
)

// GPIO pin assignments for UART0.
//
// UART0 TX is on GPIO pin 16 and RX on GPIO pin 17 (IOF0 function).
// The Nuclei UX600 variant of the SiFive GPIO block places the IOF registers
// at offsets 0x44 (IOF_EN) and 0x48 (IOF_SEL), which differ from the SiFive
// FE310/FU540 standard offsets 0x38/0x3C. The GPIO instance above is
// configured with the Nuclei offsets; see soc/sifive/gpio for details.
const (
	GPIO_UART0_TX = 16 // GPIO pin for UART0 TX (IOF0)
	GPIO_UART0_RX = 17 // GPIO pin for UART0 RX (IOF0)

	// LED shift-register chain (74HC164): GPIO 12 = CLK, GPIO 14 = DATA
	GPIO_LED_CLK  = 12 // GPIO pin for LED shift-register clock
	GPIO_LED_DATA = 14 // GPIO pin for LED shift-register data
)

// InitUARTPinmux configures the GPIO IOF registers to route UART0 signals to
// physical pins 16 (TX) and 17 (RX). IOF0 is selected for both pins (UART0
// is the IOF0 function on these pins) and the IOF override is enabled.
// Must be called during board initialization before UART0 can be used.
func InitUARTPinmux() {
	GPIO.SetIOF(GPIO_UART0_TX, 0) // pin 16 → IOF0 (UART0 TX)
	GPIO.SetIOF(GPIO_UART0_RX, 0) // pin 17 → IOF0 (UART0 RX)
}

// Model returns the SoC model name.
func Model() string {
	return "FSL91030"
}
