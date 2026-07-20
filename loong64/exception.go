// LoongArch 64-bit processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package loong64

import (
	"unsafe"

	"github.com/usbarmory/tamago/internal/exception"
)

// LoongArch exception codes (ESTAT.Ecode), see the LoongArch Reference Manual
// Volume 1, Table of Exceptions.
const (
	INT = 0x0  // interrupt
	PIL = 0x1  // page invalid, load
	PIS = 0x2  // page invalid, store
	PIF = 0x3  // page invalid, fetch
	PME = 0x4  // page modification
	PNR = 0x5  // page not readable
	PNX = 0x6  // page not executable
	PPI = 0x7  // page privilege violation
	ADE = 0x8  // address error
	ALE = 0x9  // address alignment error
	BCE = 0xa  // bound check error
	SYS = 0xb  // system call
	BRK = 0xc  // breakpoint
	INE = 0xd  // instruction non-existent
	IPE = 0xe  // instruction privilege error
	FPD = 0xf  // floating point disabled
	FPE = 0x12 // floating point error
)

// defined in exception.s
func trapHandler()

// ExceptionHandler represents an exception handler function.
type ExceptionHandler func()

func vector(fn ExceptionHandler) uint64 {
	return **((**uint64)(unsafe.Pointer(&fn)))
}

// DefaultExceptionHandler handles an exception by printing the exception cause,
// return address and bad virtual address before panicking.
func DefaultExceptionHandler() {
	estat := read_estat()

	ecode := int(estat>>16) & 0x3f
	esubcode := int(estat>>22) & 0x1ff
	is := int(estat) & 0x1fff

	print("loong64 exception: ecode ", ecode, " esubcode ", esubcode, " is ", is, " badv ", int(read_badv()), "\n")
	exception.Throw(uintptr(read_era()))
}

// SystemExceptionHandler allows to override the default exception handler.
var SystemExceptionHandler = DefaultExceptionHandler

func systemException() {
	SystemExceptionHandler()
}

// SetExceptionHandler updates the CPU exception entry with the address of the
// argument function.
func (cpu *CPU) SetExceptionHandler(fn ExceptionHandler) {
	set_eentry(vector(fn))
}

// GetExceptionHandlerAddress returns the CPU exception entry address.
func (cpu *CPU) GetExceptionHandlerAddress() uint64 {
	return vector(trapHandler)
}

// SetExceptionHandlerAddress updates the CPU exception entry address.
func (cpu *CPU) SetExceptionHandlerAddress(addr uint64) {
	set_eentry(addr)
}
