// x86-64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package amd64 provides support for AMD64 architecture specific operations.
//
// The following architectures/cores are supported/tested:
//   - AMD64 (single-core)
//
// This package is only meant to be used with `GOOS=tamago GOARCH=amd64` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package amd64

import (
	"math"
	"runtime"
	_ "unsafe"

	"github.com/usbarmory/tamago/amd64/lapic"
	"github.com/usbarmory/tamago/internal/reg"
)

// Peripheral registers
const (
	// Keyboard controller port
	KBD_PORT = 0x64
)

//go:linkname ramStackOffset runtime.ramStackOffset
var ramStackOffset uint64 = 0x100000 // 1 MB

// CPU instance
type CPU struct {
	// features
	invariant bool
	kvm       bool
	kvmclock  uint32

	// core frequency in Hz
	freq uint32
	// Timer multiplier
	TimerMultiplier float64
	// Timer offset in nanoseconds
	TimerOffset int64

	// LAPIC represents the Local APIC instance
	LAPIC *lapic.LAPIC
}

// defined in amd64.s
func exit(int32)
func halt()

// Fault generates a triple fault.
func Fault()

// Init performs initialization of an AMD64 core instance.
func (cpu *CPU) Init() {
	runtime.Exit = exit
	runtime.Idle = func(pollUntil int64) {
		// we have nothing to do forever
		if pollUntil == math.MaxInt64 {
			halt()
		}
	}

	cpu.initFeatures()
	cpu.initTimers()
}

// Name returns the CPU identifier.
func (cpu *CPU) Name() string {
	return runtime.CPU()
}

// Halt suspends execution until an interrupt is received.
func (cpu *CPU) Halt() {
	halt()
}

// Reset resets the CPU pin via 8042 keyboard controller pulse.
func (cpu *CPU) Reset() {
	reg.Out8(KBD_PORT, 0xfe)
}
