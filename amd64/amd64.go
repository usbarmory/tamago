// x86-64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package amd64 provides support for RISC-V architecture specific operations.
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
)

//go:linkname ramStackOffset runtime.ramStackOffset
var ramStackOffset uint64 = 0x100000 // 1 MB

// CPU instance
type CPU struct {
	// Timer multiplier
	TimerMultiplier float64
	// Timer offset in nanoseconds
	TimerOffset int64
	// Timer function
	TimerFn func() int64
}

// defined in amd64.s
func halt(int32)
// Fault generates a triple fault.
func Fault()

// Init performs initialization of an AMD64 core instance.
func (cpu *CPU) Init() {
	runtime.Exit = halt

	cpu.initTimers()
}

// Name returns the CPU identifier.
func (cpu *CPU) Name() string {
	return runtime.CPU()
}
