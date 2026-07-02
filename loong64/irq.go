// LoongArch 64-bit processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package loong64

import (
	"os"
	"os/signal"
	"syscall"
)

// IRQ_SIGNAL represents the `os/signal` used to signal and service interrupts.
const IRQ_SIGNAL = syscall.SIGTRAP

// CRMD.IE global interrupt enable bit.
const crmdIE = 1 << 2

// defined in irq.s
func idle()

// defined in exception.s
func handleInterrupt()

// EnableInterrupts unmasks IRQ interrupts globally by setting CRMD.IE.
func (cpu *CPU) EnableInterrupts() {
	write_crmd(read_crmd() | crmdIE)
}

// DisableInterrupts masks IRQ interrupts globally by clearing CRMD.IE.
func (cpu *CPU) DisableInterrupts() {
	write_crmd(read_crmd() &^ crmdIE)
}

// EnableInterrupt unmasks the argument local interrupt line in ECFG.LIE, the
// line index follows the ESTAT.IS layout (e.g. 11 for the timer interrupt).
func (cpu *CPU) EnableInterrupt(index int) {
	write_ecfg(read_ecfg() | (1 << index))
}

// DisableInterrupt masks the argument local interrupt line in ECFG.LIE.
func (cpu *CPU) DisableInterrupt(index int) {
	write_ecfg(read_ecfg() &^ (1 << index))
}

// WaitInterrupt suspends execution in low-power state until an interrupt is
// received.
func (cpu *CPU) WaitInterrupt() {
	idle()
}

// InterruptStatus returns the pending interrupt lines as reported by the
// ESTAT.IS field; each set bit corresponds to an interrupt source (e.g. bit
// [TimerInterrupt] for the constant frequency timer). It allows an interrupt
// service routine to determine which interrupt fired.
func (cpu *CPU) InterruptStatus() uint64 {
	return read_estat() & 0x1fff
}

// ServiceInterrupts puts the calling goroutine in wait state, its execution is
// resumed when an IRQ exception is received, the argument function can be set
// to service signaled interrupts.
func (cpu *CPU) ServiceInterrupts(isr func()) {
	if isr == nil {
		isr = func() {}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, IRQ_SIGNAL)

	for {
		// To avoid losing interrupts, re-enabling must happen only after
		// we are waiting.
		go cpu.EnableInterrupts()
		<-c
		isr()
	}
}
