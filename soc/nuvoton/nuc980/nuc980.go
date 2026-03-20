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
func disableInterrupts()
func irqEnableV5(spsr bool)
func irqDisableV5(spsr bool)
func fiqEnableV5(spsr bool)
func fiqDisableV5(spsr bool)

// CLK registers
//
// NUC980 Series Datasheet, p. 152 (§ 6.3.5 register map) and
// p. 153 (§ 6.3.6 CLK_HCLKEN), p. 157 (CLK_PCLKEN0), p. 173 (CLK_DIVCTL8).
const (
	CLK_BA = 0xB0000200

	// AHB clock enable register
	REG_CLK_HCLKEN = CLK_BA + 0x010

	// APB0 peripheral clock enable register
	// bit 8  = Timer0 clock enable
	// bit 16 = UART0 clock enable
	REG_PCLKEN0 = CLK_BA + 0x018

	// Timer0 eclk source mux: bits [17:16] = 0b00 selects XIN (12 MHz).
	REG_CLK_DIV8 = CLK_BA + 0x040

	HCLKEN_CRPT = uint32(1 << 23) // CRPT AHB clock enable (datasheet p. 153)
	PCLKEN0_TMR = uint32(1 << 8)  // Timer0 APB clock enable
	PCLKEN0_UA0 = uint32(1 << 16) // UART0 APB clock enable
)

// ARM processor instance
var ARM = &arm.CPU{
	TimerMultiplier:    1,
	NoVBAR:             true,
	IRQEnableOverride:  irqEnableV5,
	IRQDisableOverride: irqDisableV5,
	FIQEnableOverride:  fiqEnableV5,
	FIQDisableOverride: fiqDisableV5,
	WFIOverride:        waitInterrupt,
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
	// enable APB clocks for UART0 and Timer0
	// (CRPT AHB clock is enabled by initRNG before Hwinit1)
	reg.Or(REG_PCLKEN0, PCLKEN0_UA0|PCLKEN0_TMR)

	// force Timer0 eclk source to XIN (12 MHz): bits [17:16] = 0b00
	reg.Write(REG_CLK_DIV8, reg.Read(REG_CLK_DIV8)&^(uint32(0x3)<<16))

	ARM.Init(ramStart)

	// replace ARMv6+/ARMv7 exit/idle hooks with ARMv5-safe equivalents
	goos.Exit = func(_ int32) {
		disableInterrupts()
		for {
			waitInterrupt()
		}
	}
	goos.Idle = func(_ int64) {
		waitInterrupt()
	}

	// initialize serial console
	UART0.Init()

	// initialize ETimer0 as 1 MHz free-running counter for nanotime
	initTimer()
}
