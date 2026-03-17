// ARM processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm

import (
	"os"
	"os/signal"
	"syscall"
)

var irqSignal = syscall.SIGTRAP

// defined in irq.s
func irq_enable(spsr bool)
func irq_disable(spsr bool)
func fiq_enable(spsr bool)
func fiq_disable(spsr bool)
func wfi()

// EnableInterrupts unmasks IRQ interrupts in the current or saved program
// status.
func (cpu *CPU) EnableInterrupts(saved bool) {
	irq_enable(saved)
}

// DisableInterrupts masks IRQ interrupts in the current or saved program
// status.
func (cpu *CPU) DisableInterrupts(saved bool) {
	irq_disable(saved)
}

// EnableFastInterrupts unmasks FIQ interrupts in the current or saved program
// status.
func (cpu *CPU) EnableFastInterrupts(saved bool) {
	fiq_enable(saved)
}

// DisableFastInterrupts masks FIQ interrupts in the current or saved program
// status.
func (cpu *CPU) DisableFastInterrupts(saved bool) {
	fiq_disable(saved)
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
		go irq_enable(false)
		<-c
		isr()
	}
}
