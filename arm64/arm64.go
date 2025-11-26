// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package arm64 provides support for ARM 64-bit architecture specific
// operations.
//
// The following architectures/cores are supported/tested:
//   - ARMv8-A / Cortex-A53 (single-core)
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm64` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package arm64

import (
	"math"
	"runtime"
)

// CPU instance
type CPU struct {
	// Timer multiplier
	TimerMultiplier float64
	// Timer offset in nanoseconds
	TimerOffset int64
}

// defined in arm64.s
func exit(int32)

// DefaultIdleGovernor is the default CPU idle time management function
func (cpu *CPU) DefaultIdleGovernor(pollUntil int64) {
	// we have nothing to do forever
	if pollUntil == math.MaxInt64 {
		cpu.WaitInterrupt()
	}
}

// Init performs initialization of an ARM64 core instance.
func (cpu *CPU) Init() {
	runtime.Exit = exit
	runtime.Idle = cpu.DefaultIdleGovernor

	cpu.initVectorTable()
}
