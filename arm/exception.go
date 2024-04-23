// ARM processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm

import (
	"unsafe"

	"github.com/usbarmory/tamago/internal/reg"
)

// ARM exception vector offsets
// (Table 11-1, ARM® Cortex™ -A Series Programmer’s Guide).
const (
	RESET          = 0x00
	UNDEFINED      = 0x04
	SUPERVISOR     = 0x08
	PREFETCH_ABORT = 0x0c
	DATA_ABORT     = 0x10
	IRQ            = 0x18
	FIQ            = 0x1c
)

// set by application or, if not previously defined, by cpu.Init()
var vecTableStart uint32

const (
	vecTableJump   = 0xe59ff018 // ldr pc, [pc, #24]
	vecTableSize   = 0x4000     // 16 kB
	excStackOffset = 0x8000     // 32 kB
	excStackSize   = 0x4000     // 16 kB
)

// defined in exception.s
func set_exc_stack(addr uint32)
func set_vbar(addr uint32)
func set_mvbar(addr uint32)
func resetHandler()
func undefinedHandler()
func supervisorHandler()
func prefetchAbortHandler()
func dataAbortHandler()
func irqHandler()
func fiqHandler()
func nullHandler()

type ExceptionHandler func()

func vector(fn ExceptionHandler) uint32 {
	return **((**uint32)(unsafe.Pointer(&fn)))
}

type VectorTable struct {
	Reset         ExceptionHandler
	Undefined     ExceptionHandler
	Supervisor    ExceptionHandler
	PrefetchAbort ExceptionHandler
	DataAbort     ExceptionHandler
	IRQ           ExceptionHandler
	FIQ           ExceptionHandler
}

// DefaultExceptionHandler handles an exception by printing its vector and
// processor mode before panicking.
func DefaultExceptionHandler(off int) {
	print("exception: vector ", off, " mode ", int(read_cpsr()&0x1f), "\n")
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

// SystemVectorTable returns a vector table that, for all exceptions, switches
// to system mode and calls the SystemExceptionHandler on the Go runtime stack
// within goroutine g0.
func SystemVectorTable() VectorTable {
	return VectorTable{
		Reset:         resetHandler,
		Undefined:     undefinedHandler,
		Supervisor:    supervisorHandler,
		PrefetchAbort: prefetchAbortHandler,
		DataAbort:     dataAbortHandler,
		IRQ:           irqHandler,
		FIQ:           fiqHandler,
	}
}

// VectorName returns the exception vector offset name.
func VectorName(off int) string {
	switch off {
	case RESET:
		return "RESET"
	case UNDEFINED:
		return "UNDEFINED"
	case SUPERVISOR:
		return "SUPERVISOR"
	case PREFETCH_ABORT:
		return "PREFETCH_ABORT"
	case DATA_ABORT:
		return "DATA_ABORT"
	case IRQ:
		return "IRQ"
	case FIQ:
		return "FIQ"
	}

	return "Unknown"
}

// SetVectorTable updates the CPU exception handling vector table with the
// addresses of the functions defined in the passed structure.
func (cpu *CPU) SetVectorTable(t VectorTable) {
	vecTable := cpu.vbar + 8*4

	// set handler pointers
	// Table 11-1 ARM® Cortex™ -A Series Programmer’s Guide

	reg.Write(vecTable+RESET, vector(t.Reset))
	reg.Write(vecTable+UNDEFINED, vector(t.Undefined))
	reg.Write(vecTable+SUPERVISOR, vector(t.Supervisor))
	reg.Write(vecTable+PREFETCH_ABORT, vector(t.PrefetchAbort))
	reg.Write(vecTable+DATA_ABORT, vector(t.DataAbort))
	reg.Write(vecTable+IRQ, vector(t.IRQ))
	reg.Write(vecTable+FIQ, vector(t.FIQ))
}

//go:nosplit
func (cpu *CPU) initVectorTable(vbar uint32) {
	cpu.vbar = vbar

	// initialize jump table
	// Table 11-1 ARM® Cortex™ -A Series Programmer’s Guide
	for i := uint32(0); i < 8; i++ {
		reg.Write(cpu.vbar+4*i, vecTableJump)
	}

	// set exception handlers
	cpu.SetVectorTable(SystemVectorTable())

	// set vector base address register
	set_vbar(cpu.vbar)

	if cpu.Secure() {
		// set monitor vector base address register
		set_mvbar(cpu.vbar)
	}

	// Set the stack pointer for exception modes to provide a stack when
	// summoned by exception vectors.
	excStackStart := cpu.vbar + excStackOffset
	set_exc_stack(excStackStart + excStackSize)
}
