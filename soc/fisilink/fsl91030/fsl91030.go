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

	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/riscv64"
	"github.com/usbarmory/tamago/soc/sifive/clint"
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
	GPIO_BASE = 0x10012000

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

	// TODO: GPIO support (SiFive-compatible at 0x10012000)
	// Requires soc/sifive/gpio driver or new implementation.
	// GPIO IOF (I/O Function) configuration is required for UART0 to work;
	// see InitUARTPinmux() for the current pinmux-only implementation.

	// TODO: QSPI0/QSPI1 support (SiFive SPI0 compatible)
	// QSPI0 (0x10014000): NOR Flash controller (Infineon S25HL512T, 64 MB).
	// QSPI1 (0x10016000): NAND Flash/MMC; requires GPIO IOF configuration.

	// TODO: I2C0 support (OpenCores I2C compatible at 0x10018000)

	// TODO: Ethernet MAC support (FSL-specific xy1000_eth at 0x67800000)
	// Register layout:
	//   - DMA registers: base + 0x0 (RX/TX control, interrupts)
	//   - MAC registers: base + 0x400 (MAC control, PHY, statistics)
	// Interrupts: PLIC 10 (RX_END), 11 (RX_REQ), 12 (TX_END)
	// DMA buffers: 4 MB RX + 4 MB TX

	// TODO: Watchdog support (FSL-specific at 0x68000000)

	// TODO: System Clock Control (0xE084C000) and System Reset Control (0xE084E000)
	// Both are used for Ethernet peripheral clock and reset management.
)

// GPIO IOF (I/O Function) registers for UART pinmux
const (
	GPIO_IOF_SEL = 0x3C // IOF select register offset
	GPIO_IOF_EN  = 0x38 // IOF enable register offset

	// UART0 uses GPIO pins 16-17
	GPIO_UART0_MASK = 0x00030000
)

// InitUARTPinmux configures the GPIO I/O Function (IOF) registers to route
// UART0 signals to the physical pins. The IOF_SEL register selects IOF0 for
// the UART0 GPIO bits and IOF_EN activates the function, overriding GPIO mode.
// This must be called during board initialization before UART0 can be used.
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
