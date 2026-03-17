// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm64

import (
	"os"
	"os/signal"
	"syscall"
)

var irqSignal = syscall.SIGTRAP

// defined in irq.s
func irq_enable()
func irq_disable()
func wfi()

// EnableInterrupts unmasks IRQ interrupts.
func (cpu *CPU) EnableInterrupts() {
	irq_enable()
}

// DisableInterrupts masks IRQ interrupts.
func (cpu *CPU) DisableInterrupts() {
	irq_disable()
}

// WaitInterrupt suspends execution until an interrupt is received.
func (cpu *CPU) WaitInterrupt() {
	wfi()
}

// ServiceInterrupts puts the calling goroutine in wait state, its execution is
// resumed when an IRQ exception is received, an argument function can be set
// to service signaled interrupts (see gic package).
func ServiceInterrupts(isr func()) {
	if isr == nil {
		isr = func() { return }
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, irqSignal)

	for {
		// To avoid losing interrupts, re-enabling must happen only after we
		// are waiting.
		go irq_enable()
		<-c
		isr()
	}
}
