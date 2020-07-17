// ARM processor support
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package arm provides support for ARM architecture specific operations.
//
// The following architectures/cores are supported/tested:
//  * ARMv7-A / Cortex-A7 (single-core)
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/f-secure-foundry/tamago.
package arm

// ARM processor modes
// Table B1-1 ARM Architecture Reference Manual ARMv7-A and ARMv7-R edition
const (
	USR_MODE = 0b10000
	FIQ_MODE = 0b10001
	IRQ_MODE = 0b10010
	SVC_MODE = 0b10011
	MON_MODE = 0b10110
	ABT_MODE = 0b10111
	HYP_MODE = 0b11010
	UND_MODE = 0b11011
	SYS_MODE = 0b11111
)

// CPU instance
type CPU struct {
	// instruction sets
	arm     bool
	thumb   bool
	jazelle bool
	thumbee bool

	// extensions
	programmersModel bool
	security         bool
	mProfileModel    bool
	virtualization   bool
	genericTimer     bool

	// timer multiplier
	TimerMultiplier int64
	// timer function
	TimerFn func() int64
}

// defined in arm.s
func read_cpsr() uint32
func read_scr() uint32

// Init performs ARM processor instance initialization by detecting its
// available features.
func (cpu *CPU) Init() {
	cpu.initFeatures()
}

// Mode returns the processor mode.
func (cpu *CPU) Mode() int {
	return int(read_cpsr() & 0x1f)
}

// ModeName returns the processor mode name.
func ModeName(mode int) string {
	switch mode {
	case USR_MODE:
		return "USR"
	case FIQ_MODE:
		return "FIQ"
	case IRQ_MODE:
		return "IRQ"
	case SVC_MODE:
		return "SVC"
	case MON_MODE:
		return "MON"
	case ABT_MODE:
		return "ABT"
	case HYP_MODE:
		return "HYP"
	case UND_MODE:
		return "UND"
	case SYS_MODE:
		return "SYS"
	}

	return "Unknown"
}

// NonSecure returns whether the processor security mode is non-secure
// (SCR.NS).
func (cpu *CPU) NonSecure() bool {
	return (read_scr()&1 == 1)
}

// Secure returns whether the processor security mode is secure (!SCR.NS).
func (cpu *CPU) Secure() bool {
	return (read_scr()&1 == 0)
}
