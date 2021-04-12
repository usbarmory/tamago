// ARM processor support
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm

import (
	"runtime"
	"unsafe"

	"github.com/f-secure-foundry/tamago/internal/reg"
)

// ARM exception vector offsets
// Table 11-1 ARM® Cortex™ -A Series Programmer’s Guide
const (
	RESET          = 0x00
	UNDEFINED      = 0x04
	SUPERVISOR     = 0x08
	PREFETCH_ABORT = 0x0c
	DATA_ABORT     = 0x10
	IRQ            = 0x18
	FIQ            = 0x1c
)

const (
	vecTableJump uint32 = 0xe59ff018 // ldr pc, [pc, #24]

	vecTableOffset = 0
	vecTableSize   = 0x4000 // 16 kB

	excStackOffset = 0x8000 // 32 kB
	excStackSize   = 0x4000 // 16 kB
)

var (
	// ExceptionHandler defines the global exception handler, the function
	// is executed at any exception, in system mode on the Go runtime stack
	// within goroutine g0.
	ExceptionHandler = DefaultExceptionHandler

	// ResetHandler defines the reset exception handler, executed within
	// supervisor mode and its stack. SetExceptionHandlers() must be
	// invoked to update the vector table when changed.
	ResetHandler = resetHandler

	// UndefinedHandler defines the undefined exception handler, executed
	// within undefined mode and its stack. SetExceptionHandlers() must be
	// invoked to update the vector table when changed.
	UndefinedHandler = undefinedHandler

	// SupervisorHandler defines the supervisor call exception handler,
	// executed within supervisor mode and its stack.
	// SetExceptionHandlers() must be invoked to update the vector table
	// when changed.
	SupervisorHandler = svcHandler

	// PrefetchAbortHandler defines the prefetch abort exception handler,
	// executed within abort mode and its stack. SetExceptionHandlers()
	// must be invoked to update the vector table when changed.
	PrefetchAbortHandler = prefetchAbortHandler

	// DataAbortHandler defines the data abort exception handler, executed
	// within abort mode and its stack. SetExceptionHandlers() must be
	// invoked to update the vector table when changed.
	DataAbortHandler = dataAbortHandler

	// IRQHandler defines the IRQ interrupt exception handler, executed
	// within IRQ mode and its stack. SetExceptionHandlers() must be
	// invoked to update the vector table when changed.
	IRQHandler = irqHandler

	// FIQHandler defines the FIQ interrupt exception handler, executed
	// within FIQ mode and its stack. SetExceptionHandlers() must be
	// invoked to update the vector table when changed.
	FIQHandler = fiqHandler
)

// defined in exception.s
func set_exc_stack(addr uint32)
func set_vbar(addr uint32)
func resetHandler()
func undefinedHandler()
func svcHandler()
func prefetchAbortHandler()
func dataAbortHandler()
func irqHandler()
func fiqHandler()

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

func DefaultExceptionHandler(off int) {
	print("exception: vector ", off, " mode ", int(read_cpsr()&0x1f), "\n")
	panic("unhandled exception")
}

func systemException(off int) {
	ExceptionHandler(off)
}

func fnAddress(fn func()) uint32 {
	return **((**uint32)(unsafe.Pointer(&fn)))
}

// SetExceptionHandlers updates the exception handling vector table with the
// functions defined in the related global variables. It must be invoked
// whenever handling functions are changed.
func SetExceptionHandlers() {
	ramStart, _ := runtime.MemRegion()
	vecTableStart := ramStart + vecTableOffset

	// end of jump entries
	off := vecTableStart + 8*4

	// set handler pointers
	// Table 11-1 ARM® Cortex™ -A Series Programmer’s Guide

	reg.Write(off+0*4, fnAddress(ResetHandler))
	reg.Write(off+1*4, fnAddress(UndefinedHandler))
	reg.Write(off+2*4, fnAddress(SupervisorHandler))
	reg.Write(off+3*4, fnAddress(PrefetchAbortHandler))
	reg.Write(off+4*4, fnAddress(DataAbortHandler))
	reg.Write(off+5*4, fnAddress(IRQHandler))
	reg.Write(off+6*4, fnAddress(FIQHandler))
}

//go:nosplit
func (cpu *CPU) initVectorTable() {
	ramStart, _ := runtime.MemRegion()
	vecTableStart := ramStart + vecTableOffset

	// initialize jump table
	// Table 11-1 ARM® Cortex™ -A Series Programmer’s Guide
	for i := uint32(0); i < 8; i++ {
		reg.Write(vecTableStart+4*i, vecTableJump)
	}

	// set exception handlers
	SetExceptionHandlers()

	// set vector base address register
	set_vbar(vecTableStart)

	// Set the stack pointer for exception modes to provide a stack when
	// summoned by exception vectors.
	excStackStart := ramStart + excStackOffset
	set_exc_stack(excStackStart + excStackSize)
}
