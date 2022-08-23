// RISC-V processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package riscv

import (
	"unsafe"
)

// defined in exception.s
func set_mtvec(addr uint64)
func read_mepc() uint64
func read_mcause() uint64

type ExceptionHandler func()

func vector(fn ExceptionHandler) uint64 {
	return **((**uint64)(unsafe.Pointer(&fn)))
}

// DefaultExceptionHandler handles an exception by printing the exception
// program counter and trap cause before panicking.
func DefaultExceptionHandler() {
	mcause := read_mcause()
	size := XLEN - 1

	irq := int(mcause >> size)
	code := int(mcause) & ^(1 << size)

	print("exception: pc:", int(read_mepc()), " interrupt:", irq, " code:", code, "\n")
	panic("unhandled exception")
}

//go:nosplit
func (cpu *CPU) initExceptionHandler() {
	set_mtvec(vector(DefaultExceptionHandler))
}

// SetExceptionHandler updates the CPU trap vector vector with the address of
// the argument function.
func (cpu *CPU) SetExceptionHandler(fn ExceptionHandler) {
	set_mtvec(vector(fn))
}
