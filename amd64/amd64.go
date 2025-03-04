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
	"runtime"
	_ "unsafe"

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
	// Timer function
	TimerFn func() uint64
}

// defined in amd64.s
func halt(int32)

// Fault generates a triple fault.
func Fault()

// Init performs initialization of an AMD64 core instance.
func (cpu *CPU) Init() {
	runtime.Exit = halt

	cpu.initFeatures()
	cpu.initTimers()
}

// Name returns the CPU identifier.
func (cpu *CPU) Name() string {
	return runtime.CPU()
}

// Reset resets the CPU pin via 8042 keyboard controller pulse.
func (cpu *CPU) Reset() {
	reg.Out8(KBD_PORT, 0xfe)
}
