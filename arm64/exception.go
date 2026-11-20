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

// Exception Class
// Table D12.2.36 ARM Architecture Reference Manual ARMv8
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
var SystemExceptionHandler = DefaultExceptionHandler

func systemException(pc uintptr) {
	SystemExceptionHandler(pc)
}

func addJump(addr uint64, fn ExceptionHandler) {
	reg.Write64(addr, vecTableLoad)
	reg.Write64(addr+4, vecTableJump)
	reg.Write64(addr+8, fn.vector())
}

func addJumps(addr uint64, errHandler, irqHandler ExceptionHandler) {
	// Synchronous Exception
	addJump(addr, errHandler)

	// IRQ or vIRQ
	addr += 0x80
	addJump(addr, irqHandler)

	// FIQ or vFIQ
	addr += 0x80
	addJump(addr, irqHandler)

	// SError or vSError
	addr += 0x80
	addJump(addr, errHandler)
}

//go:nosplit
func (cpu *CPU) initVectorTable() {
	// 2048-bytes alignment is required
	vectorTable := uint64(goos.RamStart)

	// initialize jump tables
	// Table D1-7 ARM Architecture Reference Manual ARMv8

	// EL0
	addJumps(vectorTable, handleException, handleInterrupt)
	// ELx, x>0
	addJumps(vectorTable+0x200, handleException, handleInterrupt)

	// set vector base address register
	set_vbar(vectorTable)
}

// SetLowerVectorTable updates the CPU exception handling vector table, for
// lower exception levels.
func (cpu *CPU) SetLowerVectorTable(errHandler, irqHandler ExceptionHandler) {
	vectorTable := uint64(goos.RamStart)
	addJumps(vectorTable+0x400, errHandler, irqHandler)
}

// CurrentEL returns the Current Exception Level.
func (cpu *CPU) CurrentEL() int {
	return int(read_el()&0b1100) >> 2
}
