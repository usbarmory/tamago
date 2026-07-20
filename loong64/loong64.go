// LoongArch 64-bit processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package loong64 provides support for LoongArch 64-bit architecture specific
// operations.
//
// The following architectures/cores are supported/tested:
//   - LA64 (single-core)
//
// This package is only meant to be used with `GOOS=tamago GOARCH=loong64` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package loong64

import (
	"math"
	"runtime/goos"
	"sync"
)

// This package supports 64-bit cores.
const XLEN = 64

// CPU instance
type CPU struct {
	sync.Mutex

	// Counter represents the function to obtain the system counter for
	// operation of [CPU.GetTime] and [CPU.SetTime].
	Counter func() uint64

	// Timer multiplier
	TimerMultiplier float64
	// Timer offset in nanoseconds
	TimerOffset int64
}

// defined in loong64.s
func exit(int32)

// DefaultIdleGovernor is the default CPU idle time management function.
func (cpu *CPU) DefaultIdleGovernor(pollUntil int64) {
	// we have nothing to do forever
	if pollUntil == math.MaxInt64 {
		exit(0)
	}
}

// Init performs initialization of an LA64 core instance.
func (cpu *CPU) Init() {
	goos.Exit = exit
	goos.Idle = cpu.DefaultIdleGovernor

	if cpu.Counter == nil {
		cpu.Counter = Rdtime
	}

	cpu.SetExceptionHandler(trapHandler)
	cpu.initTimers()
}
