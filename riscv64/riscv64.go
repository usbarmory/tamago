// RISC-V 64-bit processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package riscv64 provides support for RISC-V 64-bit architecture specific
// operations.
//
// The following architectures/cores are supported/tested:
//   - RV64 (single-core)
//
// This package is only meant to be used with `GOOS=tamago GOARCH=riscv64` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package riscv64

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

// defined in riscv64.s
func exit(int32)

// DefaultIdleGovernor is the default CPU idle time management function.
// When the scheduler has no work remaining (pollUntil == math.MaxInt64) it
// executes WFI (Wait For Interrupt) to suspend the core until the next
// interrupt. Interrupts must be enabled for WFI to return.
func (cpu *CPU) DefaultIdleGovernor(pollUntil int64) {
	if pollUntil == math.MaxInt64 {
		wfi()
	}
}

// Init performs initialization of an RV64 core instance in machine mode.
func (cpu *CPU) Init() {
	goos.Exit = exit
	goos.Idle = cpu.DefaultIdleGovernor

	cpu.SetExceptionHandler(DefaultExceptionHandler)

	if cpu.Counter == nil {
		cpu.Counter = func() uint64 { return 0 }
	}
}

// InitSupervisor performs initialization of an RV64 core instance in
// supervisor mode.
func (cpu *CPU) InitSupervisor() {
	cpu.SetSupervisorExceptionHandler(DefaultSupervisorExceptionHandler)
}
