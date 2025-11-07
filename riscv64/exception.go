// RISC-V 64-bit processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package riscv64

import (
	"unsafe"

	"github.com/usbarmory/tamago/internal/exception"
)

// RISC-V exception codes (non-interrupt)
// (Table 3.6 - Volume II: RISC-V Privileged Architectures V20211203).
const (
	InstructionAddressMisaligned = 0
	InstructionAccessFault       = 1
	IllegalInstruction           = 2
	Breakpoint                   = 3
	LoadAddressMisaligned        = 4
	LoadAccessFault              = 5
	StoreAddressMisaligned       = 6
	StoreAccessFault             = 7
	EnvironmentCallFromU         = 8
	EnvironmentCallFromS         = 9
	EnvironmentCallFromM         = 11
	InstructionPageFault         = 12
	LoadPageFault                = 13
	StorePageFault               = 15
)

// defined in exception.s
func set_mtvec(addr uint64)
func set_stvec(addr uint64)
func read_mepc() uint64
func read_sepc() uint64
func read_mcause() uint64
func read_scause() uint64

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

	print("machine exception: interrupt ", irq, " code ", code, "\n")
	exception.Throw(uintptr(read_mepc()))
}

// DefaultSupervisorExceptionHandler handles an exception by printing the
// exception program counter and trap cause before panicking.
func DefaultSupervisorExceptionHandler() {
	scause := read_scause()
	size := XLEN - 1

	irq := int(scause >> size)
	code := int(scause) & ^(1 << size)

	print("supervisor exception: pc ", int(read_sepc()), " interrupt ", irq, " code ", code, "\n")
	panic("unhandled exception")
}

// SetExceptionHandler updates the CPU machine trap vector vector with the
// address of the argument function.
func (cpu *CPU) SetExceptionHandler(fn ExceptionHandler) {
	set_mtvec(vector(fn))
}

// SetSupervisorExceptionHandler updates the CPU supervisor trap vector vector
// with the address of the argument function.
func (cpu *CPU) SetSupervisorExceptionHandler(fn ExceptionHandler) {
	set_stvec(vector(fn))
}
