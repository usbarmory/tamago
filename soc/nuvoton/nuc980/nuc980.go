// Nuvoton NUC980 SoC support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package nuc980 provides support to Go bare metal unikernels written using
// the TamaGo framework on the Nuvoton NUC980 SoC (ARM926EJ-S / ARMv5TEJ).
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm GOARM=5`
// as supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package nuc980

import (
	_ "unsafe"

	"runtime/goos"

	"github.com/usbarmory/tamago/arm"
	"github.com/usbarmory/tamago/internal/reg"
)

// defined in arm926ej.s
func waitInterrupt()
func readMIDR() uint32
func disableInterrupts()

// defined in earlyinit.s
func EarlyInit()

// System registers (Global Control Register block).
//
// Confirmed against NUC980 U-Boot arch/arm/cpu/arm926ejs/nuc980/reset.c.
const (
	// SYS_PDID is the Part Device Identification register (read-only).
	// Returns the chip part number, e.g. 0x00098000 for NUC980.
	SYS_PDID = SYS_BA + 0x000

	// SYS_REGWPCTL is the register write-protection control register.
	// Writing the three-byte unlock sequence {0x59, 0x16, 0x88} in order
	// disables write protection; bit 0 reads 1 when unlocked.
	SYS_REGWPCTL = SYS_BA + 0x1FC

	// SYS_AHBIPRST is the AHB IP reset control register.
	// Writing bit 0 (CHIPRST) triggers an immediate chip reset.
	SYS_AHBIPRST = SYS_BA + 0x060
)

// PDID returns the NUC980 Part Device Identification register value.
func PDID() uint32 {
	return reg.Read(SYS_PDID)
}

// MIDR returns the ARM Main ID Register (CP15 c0,c0,0).
// Valid on all ARM cores from ARMv4 onward.
func MIDR() uint32 {
	return readMIDR()
}

// SoftReset triggers an immediate chip reset via the AHB IP reset register.
// The function does not return; the CPU restarts from the boot ROM.
func SoftReset() {
	// Unlock SYS register write protection.
	for reg.Read(SYS_REGWPCTL)&1 == 0 {
		reg.Write(SYS_REGWPCTL, 0x59)
		reg.Write(SYS_REGWPCTL, 0x16)
		reg.Write(SYS_REGWPCTL, 0x88)
	}
	// Trigger chip reset (bit 0 = CHIPRST).
	reg.Write(SYS_AHBIPRST, 0x1)
	for {
	}
}

// CLK registers
//
// NUC980 Series Datasheet, p. 152 (§ 6.3.5 register map) and
// p. 153 (§ 6.3.6 CLK_HCLKEN), p. 157 (CLK_PCLKEN0), p. 173 (CLK_DIVCTL8).
const (
	CLK_BA = 0xB0000200

	// AHB clock enable register
	REG_CLK_HCLKEN = CLK_BA + 0x010

	// APB0 peripheral clock enable register
	// bit 8  = ETimer0 clock enable
	// bit 9  = ETimer1 clock enable
	// bit 16 = UART0 clock enable
	REG_PCLKEN0 = CLK_BA + 0x018

	// Timer0 eclk source mux: bits [17:16] = 0b00 selects XIN (12 MHz).
	REG_CLK_DIV8 = CLK_BA + 0x040

	HCLKEN_CRPT  = uint32(1 << 23) // CRPT AHB clock enable (datasheet p. 153)
	HCLKEN_AIC   = uint32(1 << 10) // AIC  AHB clock enable (datasheet p. 153)
	PCLKEN0_TMR0 = uint32(1 << 8)  // ETimer0 APB clock enable
	PCLKEN0_TMR1 = uint32(1 << 9)  // ETimer1 APB clock enable
	PCLKEN0_UA0  = uint32(1 << 16) // UART0 APB clock enable
)

// ARM processor instance
var ARM = &arm.CPU{
	TimerMultiplier: 1,
	NoVBAR:          true,
}

//go:linkname ramStackOffset runtime/goos.RamStackOffset
var ramStackOffset uint32 = 0x100000 // 1 MB

//go:linkname nanotime runtime/goos.Nanotime
func nanotime() int64 {
	return int64(readTimerExtended()) * 1000
}

// Init takes care of the lower level initialization triggered early in runtime
// setup (e.g. runtime/goos.Hwinit1).
func Init() {
	// Clock gates and pin mux are configured from assembly in the
	// board cpuinit via EarlyInit (see earlyinit.s).

	ARM.Init()

	// ARM926EJ-S has no WFI opcode; override exit/idle with
	// ARMv5 equivalents.
	goos.Exit = func(_ int32) {
		disableInterrupts()
		for {
			waitInterrupt()
		}
	}
	// Start with a no-op idle governor so the scheduler can spin-poll
	// during early init.  Application code must call EnableIdleWFI()
	// after the timer interrupt infrastructure is running, otherwise
	// WFI would halt the CPU permanently.
	goos.Idle = func(_ int64) {}

	// initialize serial console
	UART0.Init()

	// initialize ETimer0 as 1 MHz free-running counter for nanotime
	initTimer()
}

// EnableIdleWFI switches the scheduler idle governor from no-op spin
// to WFI (Wait For Interrupt), reducing power consumption.  Call this
// only after a periodic hardware interrupt source (e.g. ETimer1) is
// running and arm.ServiceInterrupts is fully initialized; otherwise
// the CPU will halt permanently on the first scheduler idle.
func EnableIdleWFI() {
	goos.Idle = func(_ int64) {
		waitInterrupt()
	}
}
