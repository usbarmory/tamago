// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm64

import (
	"unsafe"
)

var (
	// set by application or, if not previously defined, by cpu.Init()
	vecTableStart uint64
	isThrowing    bool
)

const (
	vecTableJump   = 0xe59ff018 // ldr pc, [pc, #24]
	excStackOffset = 0x8000     // 32 kB
	excStackSize   = 0x4000     // 16 kB
)

// defined in exception.s
func set_vbar()

type ExceptionHandler func()

func vector(fn ExceptionHandler) uint64 {
	return **((**uint64)(unsafe.Pointer(&fn)))
}

// DefaultExceptionHandler handles an exception by printing its vector and
// processor mode before panicking.
func DefaultExceptionHandler(off int) {
	if isThrowing {
		exit(0)
	}

	// TODO: implement runtime.CallOnG0 for a cleaner approach
	isThrowing = true

	print("exception: vector ", off, "\n")
	panic("unhandled exception")
}

// SystemExceptionHandler allows to override the default exception handler
// executed at any exception by the table returned by SystemVectorTable(),
// which is used by default when initializing the CPU instance (e.g.
// CPU.Init()).
var SystemExceptionHandler = DefaultExceptionHandler

func systemException(off int) {
	SystemExceptionHandler(off)
}

//go:nosplit
func (cpu *CPU) initVectorTable() {
	// set vector base address register
	set_vbar()
}
