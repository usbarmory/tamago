// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package arm provides support for ARM architecture specific operations.
//
// The following architectures/cores are supported/tested:
//   - ARMv8-A / Cortex-A53 (single-core)
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm64` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package arm64

import (
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

// Init performs initialization of an ARM64 core instance, the argument must be
// a pointer to a 64 kB memory area which will be reserved for storing the
// exception vector table, L1/L2 page tables and the exception stack (see
// https://github.com/usbarmory/tamago/wiki/Internals#memory-layout).
func (cpu *CPU) Init(vbar uint32) {
	runtime.Exit = exit

	// the application is allowed to override the reserved area
	if vecTableStart != 0 {
		vbar = vecTableStart
	}

	// TODO
	//cpu.initVectorTable(vbar)
}
