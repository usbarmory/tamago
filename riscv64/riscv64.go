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
	"runtime"
	"sync"
)

// This package supports 64-bit cores.
const XLEN = 64

// CPU instance
type CPU struct {
	sync.Mutex
}

// defined in riscv64.s
func exit(int32)

// DefaultIdleGovernor is the default CPU idle time management function
func (cpu *CPU) DefaultIdleGovernor(pollUntil int64) {
	// we have nothing to do forever
	if pollUntil == math.MaxInt64 {
		exit(0)
	}
}

// Init performs initialization of an RV64 core instance in machine mode.
func (cpu *CPU) Init() {
	runtime.Exit = exit
	runtime.Idle = cpu.DefaultIdleGovernor

	cpu.SetExceptionHandler(DefaultExceptionHandler)
}

// InitSupervisor performs initialization of an RV64 core instance in
// supervisor mode.
func (cpu *CPU) InitSupervisor() {
	cpu.SetSupervisorExceptionHandler(DefaultSupervisorExceptionHandler)
}
