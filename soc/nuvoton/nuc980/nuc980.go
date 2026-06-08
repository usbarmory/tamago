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

// defined in init.s
func EarlyClockInit()

// System registers (Global Control Register block).
const (
	// SYS_BA is the System Manager / Global Control register base.
	SYS_BA = 0xb0000000

	// SYS_PDID is the Part Device Identification register (read-only).
	// Returns the chip part number, e.g. 0x00098000 for NUC980.
	SYS_PDID = SYS_BA + 0x000

	// SYS_REGWPCTL is the register write-protection control register.
	SYS_REGWPCTL = SYS_BA + 0x1fc

	// SYS_AHBIPRST is the AHB IP reset control register.
	SYS_AHBIPRST = SYS_BA + 0x060

	// REGWPCTL_UNLOCK reads 1 once write protection is disabled.
	REGWPCTL_UNLOCK = 0
	// AHBIPRST_CHIPRST triggers an immediate chip reset when set.
	AHBIPRST_CHIPRST = 0
)

// PDID returns the NUC980 Part Device Identification register value.
func PDID() uint32 {
	return reg.Read(SYS_PDID)
}

// MIDR returns the ARM Main ID Register.
func MIDR() uint32 {
	return ARM.MIDR()
}

// SoftReset triggers an immediate chip reset via the AHB IP reset register.
// The function does not return; the CPU restarts from the boot ROM.
func SoftReset() {
	// Unlock SYS register write protection.
	for !reg.Get(SYS_REGWPCTL, REGWPCTL_UNLOCK) {
		reg.Write(SYS_REGWPCTL, 0x59)
		reg.Write(SYS_REGWPCTL, 0x16)
		reg.Write(SYS_REGWPCTL, 0x88)
	}
	// Trigger chip reset.
	reg.Set(SYS_AHBIPRST, AHBIPRST_CHIPRST)
	select {}
}

// CLK registers
const (
	CLK_BA = 0xb0000200

	// AHB clock enable register
	REG_CLK_HCLKEN = CLK_BA + 0x010

	// APB0 peripheral clock enable register
	// bit 8  = ETimer0 clock enable
	// bit 9  = ETimer1 clock enable
	// bit 16 = UART0 clock enable
	REG_PCLKEN0 = CLK_BA + 0x018

	// Timer0/Timer1 eclk source mux: clearing bits [19:16] selects XIN.
	REG_CLK_DIV8      = CLK_BA + 0x040
	CLK_DIV8_ECLK_XIN = 0xf << 16

	HCLKEN_CRPT  = 1 << 23 // CRPT AHB clock enable
	HCLKEN_AIC   = 1 << 10 // AIC  AHB clock enable
	PCLKEN0_TMR0 = 1 << 8  // ETimer0 APB clock enable
	PCLKEN0_TMR1 = 1 << 9  // ETimer1 APB clock enable
	PCLKEN0_UA0  = 1 << 16 // UART0 APB clock enable
)

// ARM processor instance
var ARM = &arm.CPU{
	TimerMultiplier: 1,
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
	// board cpuinit via EarlyClockInit (see init.s).

	ARM.Init()

	// ARM926EJ-S has no WFI opcode; override exit/idle with
	// ARMv5 equivalents (arm.WaitInterrupt uses the CP15 WFI on
	// GOARM=5, see arm/irq_v5.s).
	goos.Exit = func(_ int32) {
		ARM.DisableInterrupts(false)
		ARM.DisableFastInterrupts(false)
		for {
			ARM.WaitInterrupt()
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

	// seed the PRNG from the free-running timer now that it is running
	PRNG.Seed = uint32(nanotime())
}

// EnableIdleWFI switches the scheduler idle governor from no-op spin
// to WFI (Wait For Interrupt), reducing power consumption.  Call this
// only after a periodic hardware interrupt source (e.g. ETimer1) is
// running and arm.ServiceInterrupts is fully initialized; otherwise
// the CPU will halt permanently on the first scheduler idle.
func EnableIdleWFI() {
	goos.Idle = func(_ int64) {
		ARM.WaitInterrupt()
	}
}
