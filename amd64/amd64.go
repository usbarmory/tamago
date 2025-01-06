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

import "runtime"

// CPU instance
type CPU struct{}

// defined in amd64.s
func Fault()
func halt(int32)

// Init performs initialization of an AMD64 core instance.
func (cpu *CPU) Init() {
	runtime.Exit = halt
}
