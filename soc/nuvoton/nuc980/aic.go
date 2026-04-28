// Nuvoton NUC980 Advanced Interrupt Controller (AIC) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Minimal AIC driver for the NUC980 SoC.
//
// The AIC supports 64 interrupt sources (IRQ0..IRQ63).  Sources 0..31 are
// controlled by MECR/MDCR (low word); sources 32..63 by MECRH/MDCRH.
//
// Register references: NUC980 Series Datasheet, p. 180 (§ 6.4 Advanced
// Interrupt Controller).

package nuc980

import (
	"github.com/usbarmory/tamago/internal/reg"
)

// AIC register addresses
//
// NUC980 Series Datasheet, p. 180 (§ 6.4 Advanced Interrupt Controller).
const (
	AIC_BA = 0xB0042000

	// Software Interrupt Set Command Register
	REG_AIC_SISCR = AIC_BA + 0x100

	// Software Interrupt Clear Command Register
	REG_AIC_SICCR = AIC_BA + 0x104

	// Interrupt source number register (current IRQ)
	REG_AIC_ISNR = AIC_BA + 0x120

	// Mask enable (set bits enable the corresponding IRQ sources)
	REG_AIC_MECR  = AIC_BA + 0x130 // IRQ0..31
	REG_AIC_MECRH = AIC_BA + 0x134 // IRQ32..63

	// Mask disable (set bits disable the corresponding IRQ sources)
	REG_AIC_MDCR  = AIC_BA + 0x138 // IRQ0..31
	REG_AIC_MDCRH = AIC_BA + 0x13C // IRQ32..63

	// End-of-Service Command Register (write any value to signal EOI)
	REG_AIC_EOSCR = AIC_BA + 0x150
)

// AIC interrupt source numbers.
//
// Linux: arch/arm/mach-nuc980/include/mach/irqs.h
const (
	IRQ_ETMR0 = 16 // Enhanced Timer 0
	IRQ_ETMR1 = 17 // Enhanced Timer 1
	IRQ_UART0 = 24 // UART 0
)

// DisableAll masks all 64 AIC interrupt sources.
func DisableAll() {
	reg.Write(REG_AIC_MDCR, 0xFFFFFFFF)
	reg.Write(REG_AIC_MDCRH, 0xFFFFFFFF)
}

// EnableIRQ unmasks interrupt source irq (0..63).
func EnableIRQ(irq int) {
	if irq < 0 || irq > 63 {
		return
	}
	if irq < 32 {
		reg.Write(REG_AIC_MECR, 1<<uint(irq))
	} else {
		reg.Write(REG_AIC_MECRH, 1<<uint(irq-32))
	}
}

// DisableIRQ masks interrupt source irq (0..63).
func DisableIRQ(irq int) {
	if irq < 0 || irq > 63 {
		return
	}
	if irq < 32 {
		reg.Write(REG_AIC_MDCR, 1<<uint(irq))
	} else {
		reg.Write(REG_AIC_MDCRH, 1<<uint(irq-32))
	}
}

// CurrentIRQ returns the interrupt source number currently being serviced.
func CurrentIRQ() int {
	return int(reg.Read(REG_AIC_ISNR) & 0x7F)
}

// EOI signals end-of-interrupt to the AIC.
func EOI() {
	reg.Write(REG_AIC_EOSCR, 0x1)
}

// SoftwareInterrupt generates a software interrupt for the given IRQ source.
func SoftwareInterrupt(irq int) {
	if irq < 0 || irq > 63 {
		return
	}
	if irq < 32 {
		reg.Write(REG_AIC_SISCR, 1<<uint(irq))
	} else {
		// NUC980 datasheet: SISCR only covers IRQ0..31.
		// Software interrupts for IRQ32+ are not supported.
	}
}
