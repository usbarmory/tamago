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
	"fmt"
	_ "unsafe"
)

// ARM exception vector offsets
// Table 11-1 ARM® Cortex™ -A Series Programmer’s Guide
const (
	RESET          = 0x0
	UNDEFINED      = 0x04
	SUPERVISOR     = 0x08
	PREFETCH_ABORT = 0x0c
	DATA_ABORT     = 0x10
	IRQ            = 0x18
	FIQ            = 0x1c
)

var exceptionHandlerFn = defaultExceptionHandler

//go:linkname exceptionHandler runtime.exceptionHandler
func exceptionHandler(off int) {
	exceptionHandlerFn(off)
}

func defaultExceptionHandler(off int) {
	mode := int(read_cpsr() & 0x1f)
	msg := fmt.Sprintf("unhandled exception, vector %#x (%s), mode %#x (%s)", off, VectorName(off), mode, ModeName(mode))
	panic(msg)
}

// ExceptionHandler overrides the default exception handler, the passed
// function receives the exception vector offset as argument.
func ExceptionHandler(fn func(int)) {
	exceptionHandlerFn = fn
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
