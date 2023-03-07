// ARM processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm

// defined in irq.s
func irq_enable(spsr bool)
func irq_disable(spsr bool)
func fiq_enable(spsr bool)
func fiq_disable(spsr bool)

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
