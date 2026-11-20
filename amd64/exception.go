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

// SystemExceptionHandler allows to override the default exception handler.
//
// The handler executes on the stack and goroutine (g) of the interrupted code,
// with interrupts masked, and must not return. It should avoid allocations,
// locks or any runtime state which might be inconsistent at exception time.
var SystemExceptionHandler = DefaultExceptionHandler

// EnableExceptions initializes handling of processor exceptions through
// DefaultExceptionHandler().
func (cpu *CPU) EnableExceptions() {
	// Handle processor exceptions, stop before VMM Communication to
	// preserve original handler if present (e.g. under UEFI+SEV-SNP).
	setIDT(0, 28)

	// user defined interrupts
	setIDT(32, 255)
}
