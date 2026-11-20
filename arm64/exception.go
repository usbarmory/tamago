// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm64

import (
	"runtime/goos"
	"unsafe"

	"github.com/usbarmory/tamago/internal/exception"
	"github.com/usbarmory/tamago/internal/reg"
)

const (
	vecTableLoad = 0x58000052 // ldr x18, #8
	vecTableJump = 0xd61f0240 // br x18
)

// Exception classes
// (D12.2.36 ARM Architecture Reference Manual ARMv8).
const (
	ESR_EC     = 26
	EC_UNKNOWN = 0b000000
	EC_SVC64   = 0b010101
	EC_SYSREG  = 0b011000
	EC_SERROR  = 0b101111
)

// defined in exception.s
func set_vbar(addr uint64)
func read_el() uint64
func read_esr() uint64
func handleException()
func handleInterrupt()

// ExceptionHandler represents an exception vector entry point, reached through
// a trampoline which uses R18 to branch to it (its value at exception time is
// therefore lost).
//
// The handler must be an assembly function marked NOSPLIT|NOFRAME, as its
// address is taken from the function value (i.e. its ABI wrapper). Fatal
// exceptions can be handled in Go by overriding [SystemExceptionHandler].
type ExceptionHandler func()

func (fn ExceptionHandler) vector() uint64 {
	return **((**uint64)(unsafe.Pointer(&fn)))
}

var isThrowing bool

// DefaultExceptionHandler handles an exception by printing its vector and
// processor mode before panicking.
func DefaultExceptionHandler(pc uintptr) {
	if isThrowing {
		goos.Exit(1)
	}

	isThrowing = true

	el := (read_el() & 0b1100) >> 2
	ec := (read_esr() & 0b111111) >> ESR_EC

	print("EL", int(el), " exception class ", int(ec), "\n")
	exception.Throw(pc)
}

// SystemExceptionHandler allows to override the default exception handler.
//
// The handler executes on the stack and goroutine (g) of the interrupted code,
// with interrupts masked, and must not return. It should avoid allocations,
// locks or any runtime state which might be inconsistent at exception time.
var SystemExceptionHandler = DefaultExceptionHandler

func systemException(pc uintptr) {
	SystemExceptionHandler(pc)
}

func addJump(addr uint64, fn ExceptionHandler) {
	reg.Write64(addr, vecTableLoad)
	reg.Write64(addr+4, vecTableJump)
	reg.Write64(addr+8, fn.vector())
}

func addJumps(addr uint64, syncHandler, irqHandler ExceptionHandler) {
	// Synchronous Exception
	addJump(addr, syncHandler)

	// IRQ or vIRQ
	addr += 0x80
	addJump(addr, irqHandler)

	// FIQ or vFIQ
	addr += 0x80
	addJump(addr, irqHandler)

	// SError or vSError
	addr += 0x80
	addJump(addr, syncHandler)
}

//go:nosplit
func (cpu *CPU) initVectorTable() {
	// 2048-bytes alignment is required
	vectorTable := uint64(goos.RamStart)

	// initialize jump tables
	// Table D1-7 ARM Architecture Reference Manual ARMv8

	// current EL with SP_EL0
	addJumps(vectorTable, handleException, handleInterrupt)
	// current with ELx, x>0
	addJumps(vectorTable+0x200, handleException, handleInterrupt)
	// Lower EL, AArch64
	addJumps(vectorTable+0x400, handleException, handleInterrupt)

	// set vector base address register
	set_vbar(vectorTable)
}

// SetLowerVectorTable updates the CPU exception handling vector table, for
// lower exception levels.
func (cpu *CPU) SetLowerVectorTable(syncHandler, irqHandler ExceptionHandler) {
	vectorTable := uint64(goos.RamStart)
	addJumps(vectorTable+0x400, syncHandler, irqHandler)
}

// CurrentEL returns the Current Exception Level.
func (cpu *CPU) CurrentEL() int {
	return int(read_el()&0b1100) >> 2
}
