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
// The package implements initialization and drivers for specific Fisilink
// FSL91030 System-on-Chip (SoC) peripherals, based on the Nuclei UX600
// RISC-V core, adopting, where indicated, the following reference
// specifications:
//   - nuclei_ux600fd.dts - Device tree from vega-buildroot-sdk
//   - freeloader.S - Boot loader assembly from vega-loader-entire
//   - platform.c - OpenSBI platform support (vega-opensbi/platform/nuclei/ux600)
//
// The FSL91030 is a RV64IMAFDC processor with sv39 MMU, running at 400 MHz.
// Many peripherals are SiFive-compatible (UART, GPIO, SPI), while others
// (Ethernet MAC, Watchdog) are FSL-specific.
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
	// Nuclei Timer (base for CLINT-compatible timer)
	// The Nuclei UX600 has a timer at 0x2000000
	// CLINT-compatible offset is at +0x1000
	NUCLEI_TIMER_BASE = 0x2000000
	CLINT_BASE        = NUCLEI_TIMER_BASE + 0x1000

	// Platform-Level Interrupt Controller (PLIC)
	// Note: PLIC support deferred to board package
	PLIC_BASE = 0x8000000

	// DDR Controller
	DDR_BASE = 0x10001000

	// GPIO (SiFive-compatible, used for UART pinmux)
	GPIO_BASE = 0x10012000

	// Serial ports (SiFive UART0 compatible)
	UART0_BASE = 0x10013000 // Console
	UART1_BASE = 0x10023000 // Secondary UART

	// QSPI Flash Controllers (SiFive SPI0 compatible)
	QSPI0_BASE = 0x10014000 // NOR Flash
	QSPI1_BASE = 0x10016000 // NAND Flash/MMC

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
	RV64 = &riscv64.CPU{}

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

	// TODO: GPIO support (SiFive-compatible at 0x10012000)
	// Requires soc/sifive/gpio driver or new implementation
	// CRITICAL: GPIO pinmux must be configured for UART0 to work
	// See InitUARTPinmux() for required initialization

	// TODO: QSPI0/QSPI1 support (SiFive SPI0 compatible)
	// Requires SPI driver implementation
	// QSPI0: 0x10014000 - NOR Flash (Macronix MX25U51245G)
	// QSPI1: 0x10016000 - NAND Flash/MMC (for SD card boot)
	// Note: QSPI1 requires GPIO pinmux configuration for SD boot (see u-boot board file)

	// TODO: I2C0 support (OpenCores I2C compatible)
	// Requires I2C driver implementation

	// TODO: Ethernet MAC support (FSL-specific at 0x67800000)
	// Requires custom driver for xy1000_eth MAC (see vega-u-boot/drivers/net/xy1000_eth.c)
	// Register structure:
	//   - DMA registers: base + 0x0 (RX/TX control, interrupts)
	//   - MAC registers: base + 0x400 (MAC control, PHY, statistics)
	// Interrupts: PLIC 10 (RX_END), 11 (RX_REQ), 12 (TX_END)
	// DMA buffer requirements: RX 4MB, TX 4MB

	// TODO: Watchdog support (FSL-specific at 0x68000000)
	// Requires custom driver for FSL WDT

	// TODO: System Clock Control (0xE084C000)
	// Used by ethernet driver for clock configuration

	// TODO: System Reset Control (0xE084E000)
	// Used by ethernet driver for reset control
)

// GPIO IOF (I/O Function) registers for UART pinmux
const (
	GPIO_IOF_SEL = 0x3C // IOF select register offset
	GPIO_IOF_EN  = 0x38 // IOF enable register offset

	// UART0 uses GPIO pins 16-17
	GPIO_UART0_MASK = 0x00030000
)

// InitUARTPinmux configures GPIO pins for UART0 functionality.
//
// The Nuclei UX600 requires GPIO I/O Function (IOF) configuration to enable
// UART0 on the physical pins. This function must be called during board
// initialization before UART0 can be used.
//
// Based on OpenSBI platform.c:ux600_early_init()
func InitUARTPinmux() {
	// GPIO IOF_SEL: Clear UART0 bits (select IOF0)
	iofSel := reg.Read(GPIO_BASE + GPIO_IOF_SEL)
	iofSel &= ^uint32(GPIO_UART0_MASK)
	reg.Write(GPIO_BASE+GPIO_IOF_SEL, iofSel)

	// GPIO IOF_EN: Set UART0 bits (enable IOF)
	iofEn := reg.Read(GPIO_BASE + GPIO_IOF_EN)
	iofEn |= GPIO_UART0_MASK
	reg.Write(GPIO_BASE+GPIO_IOF_EN, iofEn)
}

// Model returns the SoC model name.
func Model() string {
	return "FSL91030"
}
