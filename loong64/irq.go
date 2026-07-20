// LoongArch 64-bit processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package loong64

// CRMD.IE global interrupt enable bit.
const crmdIE = 1 << 2

// defined in irq.s
func idle()

// EnableInterrupts unmasks IRQ interrupts by setting CRMD.IE.
func (cpu *CPU) EnableInterrupts() {
	write_crmd(read_crmd() | crmdIE)
}

// DisableInterrupts masks IRQ interrupts by clearing CRMD.IE.
func (cpu *CPU) DisableInterrupts() {
	write_crmd(read_crmd() &^ crmdIE)
}

// WaitInterrupt suspends execution in low-power state until an interrupt is
// received.
func (cpu *CPU) WaitInterrupt() {
	idle()
}
