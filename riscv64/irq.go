// RISC-V 64-bit interrupt support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package riscv64

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
func trapHandler()

// EnableInterrupts enables machine-mode external interrupts and installs
// the unified trap handler at mtvec.
func (cpu *CPU) EnableInterrupts() {
	cpu.SetExceptionHandler(trapHandler)
	irq_enable()
}

// DisableInterrupts disables machine-mode external interrupts.
func (cpu *CPU) DisableInterrupts() {
	irq_disable()
}

// WaitInterrupt suspends the CPU until the next interrupt (WFI).
func (cpu *CPU) WaitInterrupt() {
	wfi()
}

// ServiceInterrupts puts the calling goroutine in wait state; its execution is
// resumed when a machine-mode external interrupt is received. The isr function
// is called to service pending interrupts after each wakeup.
//
// Interrupts are re-enabled in a separate goroutine after each wakeup so that
// the channel receive is reached before the next interrupt can fire, preventing
// the race where signal.Relay is called before the goroutine is waiting.
//
// ServiceInterrupts never returns.
func (cpu *CPU) ServiceInterrupts(isr func()) {
	if isr == nil {
		isr = func() {}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, irqSignal)

	cpu.SetExceptionHandler(trapHandler)

	for {
		go irq_enable()
		<-c
		isr()
	}
}
