// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm64

import (
	"math"
	"runtime"
	"time"
)

// IRQ handling goroutine
var irqHandlerG uint

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
	irqHandlerG, _ = runtime.GetG()

	if isr == nil {
		isr = func() { return }
	}

	for {
		// To avoid losing interrupts, re-enabling must happen only after we
		// are sleeping.
		go irq_enable()

		// Sleep indefinitely until woken up by runtime.WakeG
		// (see ·handleInterrupt in irq.s).
		time.Sleep(math.MaxInt64)

		isr()
	}
}
