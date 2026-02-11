// AMD64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package amd64

import (
	"runtime/goos"

	"github.com/usbarmory/tamago/internal/exception"
)

var (
	isr        uintptr
	eip        uintptr
	isThrowing bool
)

func currentVectorNumber() (id int) {
	id = int(isr - irqHandlerAddr)

	if id >= 0 {
		id = id / callSize
	}

	return
}

// DefaultExceptionHandler handles an exception by printing its vector and
// processor mode before panicking.
func DefaultExceptionHandler() {
	if isThrowing {
		goos.Exit(1)
	}

	isThrowing = true

	print("exception: vector ", currentVectorNumber(), " \n")
	exception.Throw(eip)
}

// SystemExceptionHandler allows to override the default exception handler
// executed at any exception by the table returned by SystemVectorTable(),
// which is used by default when initializing the CPU instance (e.g.
// CPU.Init()).
var SystemExceptionHandler = DefaultExceptionHandler

// EnableExceptions initializes handling of processor exceptions through
// DefaultExceptionHandler().
func (cpu *CPU) EnableExceptions() {
	// processor exceptions
	setIDT(0, 31)
}
