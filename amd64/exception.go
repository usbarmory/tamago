// x86-64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package amd64

var (
	currentVector uintptr
	isThrowing    bool
)

func currentVectorNumber() (id int) {
	id = int(currentVector - irqHandlerAddr)

	if id >= 0 {
		id = id / callSize
	}

	return
}

// DefaultExceptionHandler handles an exception by printing its vector and
// processor mode before panicking.
func DefaultExceptionHandler() {
	if isThrowing {
		exit(0)
	}

	// TODO: implement runtime.CallOnG0 for a cleaner approach
	isThrowing = true

	print("exception: vector ", currentVectorNumber(), " \n")
	panic("unhandled exception")
}

// EnableExceptions initializes handling of processor exceptions through
// DefaultExceptionHandler().
func (cpu *CPU) EnableExceptions() {
	// processor exceptions
	setIDT(0, 31)
}
