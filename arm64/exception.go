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

// defined in exception.s
func set_vbar(addr uint64)
func read_el() uint64
func handleException()
func handleInterrupt()

type exceptionHandler func()

func (fn exceptionHandler) vector() uint64 {
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

	print("EL", int(read_el()&0b1100)>>2, " exception\n")
	exception.Throw(pc)
}

// SystemExceptionHandler allows to override the default exception handler.
var SystemExceptionHandler = DefaultExceptionHandler

func systemException(pc uintptr) {
	SystemExceptionHandler(pc)
}

func addJump(addr uint64, fn exceptionHandler) {
	reg.Write64(addr, vecTableLoad)
	reg.Write64(addr+4, vecTableJump)
	reg.Write64(addr+8, fn.vector())
}

func addJumps(addr uint64) {
	// Synchronous Exception
	addJump(addr, handleException)

	// IRQ or vIRQ
	addr += 0x80
	addJump(addr, handleInterrupt)

	// FIQ or vFIQ
	addr += 0x80
	addJump(addr, handleInterrupt)

	// SError or vSError
	addr += 0x80
	addJump(addr, handleException)
}

//go:nosplit
func (cpu *CPU) initVectorTable() {
	// 2048-bytes alignment is required
	vectorTable := uint64(goos.RamStart)

	// initialize jump tables
	// Table D1-7 ARM Architecture Reference Manual ARMv8

	// EL0
	addJumps(vectorTable)
	// ELx, x>0
	addJumps(vectorTable + 0x200)

	// set vector base address register
	set_vbar(vectorTable)
}
