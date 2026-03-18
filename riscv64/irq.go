// RISC-V 64-bit interrupt support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package riscv64

import (
	"math"
	"runtime"
	"time"
)

// irqHandlerG holds the goroutine pointer for the interrupt handler goroutine.
// Set by ServiceInterrupts; read from assembly by trapHandler via WakeG.
var irqHandlerG uint

// defined in irq.s
func irq_enable()
func irq_disable()
func wfi()
func trapHandler()

// EnableInterrupts enables machine-mode external interrupts and installs
// the unified trap handler at mtvec. Safe to call before ServiceInterrupts
// if a custom ISR pattern is used.
func (cpu *CPU) EnableInterrupts() {
	cpu.SetExceptionHandler(trapHandler)
	irq_enable()
}

// DisableInterrupts disables machine-mode external interrupts (clears mie.MEIE).
// Does not affect other interrupt sources (timer, software).
func (cpu *CPU) DisableInterrupts() {
	irq_disable()
}

// WaitInterrupt suspends the CPU until the next interrupt is received (WFI).
func (cpu *CPU) WaitInterrupt() {
	wfi()
}

// ServiceInterrupts puts the calling goroutine in wait state; its execution is
// resumed when a machine-mode external interrupt is received. The argument
// function is called to service pending interrupts after each wakeup.
//
// Interrupts are re-enabled in a separate goroutine after each wakeup so that
// the handler goroutine reaches time.Sleep before the next interrupt can fire,
// preventing the race where WakeG is called before the goroutine timer is live.
//
// The isr function must drain all pending interrupts on each call; an unclaimed
// interrupt causes the PLIC to re-assert immediately after MRET.
// ServiceInterrupts never returns.
func (cpu *CPU) ServiceInterrupts(isr func()) {
	irqHandlerG, _ = runtime.GetG()

	if isr == nil {
		isr = func() {}
	}

	cpu.SetExceptionHandler(trapHandler)

	for {
		// Re-enable interrupts only after this goroutine has entered
		// time.Sleep, so the goroutine timer is live before the first
		// interrupt can fire.
		go irq_enable()

		// Sleep until woken by runtime.WakeG from the trap handler.
		time.Sleep(math.MaxInt64)

		isr()
	}
}
