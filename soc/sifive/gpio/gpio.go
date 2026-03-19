// SiFive GPIO driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package gpio implements a driver for the SiFive GPIO0 peripheral block,
// as used in SiFive FE310, FU540, and Nuclei UX600-based SoCs.
//
// Register layout based on:
//   - SiFive FE310-G002 Manual (gpio@10012000)
//   - Linux kernel gpio-sifive.c
//   - Nuclei UX600 OpenSBI platform code (ux600/platform.c)
//
// NOTE on IOF register offsets: The Nuclei UX600 variant places IOF_EN at
// offset 0x44 and IOF_SEL at 0x48. Standard SiFive FE310/FU540 use 0x38/0x3C.
// Set the IOFENOffset and IOFSELOffset fields accordingly when constructing a
// GPIO instance for Nuclei UX600 hardware.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=riscv64` as
// supported by the TamaGo framework for bare metal Go on RISC-V SoCs, see
// https://github.com/usbarmory/tamago.
package gpio

import (
	"github.com/usbarmory/tamago/internal/reg"
)

// GPIO register offsets for the Nuclei UX600/UX608 variant (FSL91030).
//
// This layout is confirmed by the FSL91030M Register Specification (Section 6.1)
// and the Linux gpio-sifive.c driver from the Vega buildroot SDK. It differs
// from the standard SiFive FE310/FU540 layout: the Nuclei variant inserts three
// extra registers (PDE, OPEN_DRAIN, PUP) between DS and RISE_IE, shifting all
// interrupt registers and later registers down by 0x0C.
//
// Standard SiFive FE310/FU540 has:  RISE_IE=0x18, IOF_EN=0x38, OUT_XOR=0x40
// Nuclei UX600/UX608 (FSL91030) has: RISE_IE=0x24, IOF_EN=0x44, OUT_XOR=0x58
const (
	GPIO_INPUT_VAL  = 0x00 // Input value (read pin state)
	GPIO_INPUT_EN   = 0x04 // Input enable
	GPIO_OUTPUT_EN  = 0x08 // Output enable
	GPIO_OUTPUT_VAL = 0x0C // Output value
	GPIO_PUE        = 0x10 // Pull-up enable
	GPIO_DS         = 0x14 // Drive strength
	GPIO_PDE        = 0x18 // Pull-down enable (Nuclei UX600 addition)
	GPIO_OPEN_DRAIN = 0x1C // Open-drain enable (Nuclei UX600 addition)
	GPIO_PUP        = 0x20 // Push-pull enable (Nuclei UX600 addition)
	GPIO_RISE_IE    = 0x24 // Rise interrupt enable
	GPIO_RISE_IP    = 0x28 // Rise interrupt pending (write-1-to-clear)
	GPIO_FALL_IE    = 0x2C // Fall interrupt enable
	GPIO_FALL_IP    = 0x30 // Fall interrupt pending (write-1-to-clear)
	GPIO_HIGH_IE    = 0x34 // High-level interrupt enable
	GPIO_HIGH_IP    = 0x38 // High-level interrupt pending
	GPIO_LOW_IE     = 0x3C // Low-level interrupt enable
	GPIO_LOW_IP     = 0x40 // Low-level interrupt pending

	// IOF (I/O Function) registers.
	// Nuclei UX600/UX608 (FSL91030): IOF_EN=0x44, IOF_SEL0=0x48, IOF_SEL1=0x4C
	// Standard SiFive FE310/FU540:   IOF_EN=0x38, IOF_SEL=0x3C (no SEL1)
	// Set IOFENOffset / IOFSELOffset on the GPIO struct to select.
	GPIO_IOF_EN_SIFIVE  = 0x38 // IOF enable (SiFive FE310/FU540)
	GPIO_IOF_SEL_SIFIVE = 0x3C // IOF select (SiFive FE310/FU540)
	GPIO_IOF_EN_NUCLEI  = 0x44 // IOF enable (Nuclei UX600/UX608 / FSL91030)
	GPIO_IOF_SEL_NUCLEI = 0x48 // IOF function select 0 (Nuclei UX600/UX608)
	GPIO_IOF_SEL1       = 0x4C // IOF function select 1 (Nuclei UX600/UX608 only)

	GPIO_EVENT_RISE_EN = 0x50 // Rising-edge event enable (Nuclei UX600 addition)
	GPIO_EVENT_FALL_EN = 0x54 // Falling-edge event enable (Nuclei UX600 addition)
	GPIO_OUTPUT_XOR    = 0x58 // Output XOR (invert output)
	GPIO_SW_FILTER_EN  = 0x5C // Schmitt-trigger input filter enable (Nuclei UX600)

	// GPIO_MAX_PINS is the maximum number of pins per GPIO block.
	GPIO_MAX_PINS = 32
)

// GPIO represents a SiFive GPIO0 peripheral instance.
//
// Create an instance and set Base to the MMIO base address. For Nuclei UX600
// SoCs (e.g. FSL91030), also set IOFENOffset and IOFSELOffset:
//
//	GPIO = &gpio.GPIO{
//	    Base:         0x10011000,
//	    IOFENOffset:  gpio.GPIO_IOF_EN_NUCLEI,
//	    IOFSELOffset: gpio.GPIO_IOF_SEL_NUCLEI,
//	}
type GPIO struct {
	// Base is the MMIO base address of the GPIO block.
	Base uint32

	// IOFENOffset is the byte offset of the IOF_EN register from Base.
	// Use GPIO_IOF_EN_SIFIVE (0x38) for SiFive FE310/FU540.
	// Use GPIO_IOF_EN_NUCLEI (0x44) for Nuclei UX600 (e.g. FSL91030).
	// Defaults to GPIO_IOF_EN_SIFIVE when zero.
	IOFENOffset uint32

	// IOFSELOffset is the byte offset of the IOF_SEL register from Base.
	// Use GPIO_IOF_SEL_SIFIVE (0x3C) for SiFive FE310/FU540.
	// Use GPIO_IOF_SEL_NUCLEI (0x48) for Nuclei UX600 (e.g. FSL91030).
	// Defaults to GPIO_IOF_SEL_SIFIVE when zero.
	IOFSELOffset uint32
}

// iofENAddr returns the effective IOF_EN register address.
func (hw *GPIO) iofENAddr() uint32 {
	if hw.IOFENOffset == 0 {
		return hw.Base + GPIO_IOF_EN_SIFIVE
	}
	return hw.Base + hw.IOFENOffset
}

// iofSELAddr returns the effective IOF_SEL register address.
func (hw *GPIO) iofSELAddr() uint32 {
	if hw.IOFSELOffset == 0 {
		return hw.Base + GPIO_IOF_SEL_SIFIVE
	}
	return hw.Base + hw.IOFSELOffset
}

// pinMask returns a 32-bit mask with bit n set. Panics if n is out of range.
func pinMask(n int) uint32 {
	if n < 0 || n >= GPIO_MAX_PINS {
		panic("gpio: pin number out of range")
	}
	return 1 << uint(n)
}

// SetOutputEnable configures pin n as an output.
func (hw *GPIO) SetOutputEnable(n int) {
	reg.Set(hw.Base+GPIO_OUTPUT_EN, n)
}

// ClearOutputEnable configures pin n as an input (disables output driver).
func (hw *GPIO) ClearOutputEnable(n int) {
	reg.Clear(hw.Base+GPIO_OUTPUT_EN, n)
}

// SetInputEnable enables the input buffer on pin n.
func (hw *GPIO) SetInputEnable(n int) {
	reg.Set(hw.Base+GPIO_INPUT_EN, n)
}

// ClearInputEnable disables the input buffer on pin n.
func (hw *GPIO) ClearInputEnable(n int) {
	reg.Clear(hw.Base+GPIO_INPUT_EN, n)
}

// Out drives pin n high.
func (hw *GPIO) Out(n int) {
	reg.Set(hw.Base+GPIO_OUTPUT_VAL, n)
}

// Clear drives pin n low.
func (hw *GPIO) Clear(n int) {
	reg.Clear(hw.Base+GPIO_OUTPUT_VAL, n)
}

// Toggle inverts the current output state of pin n.
func (hw *GPIO) Toggle(n int) {
	val := reg.Read(hw.Base + GPIO_OUTPUT_VAL)
	reg.Write(hw.Base+GPIO_OUTPUT_VAL, val^pinMask(n))
}

// In returns true if pin n input reads high.
func (hw *GPIO) In(n int) bool {
	return reg.Read(hw.Base+GPIO_INPUT_VAL)&pinMask(n) != 0
}

// SetPullUp enables the pull-up resistor on pin n.
func (hw *GPIO) SetPullUp(n int) {
	reg.Set(hw.Base+GPIO_PUE, n)
}

// ClearPullUp disables the pull-up resistor on pin n.
func (hw *GPIO) ClearPullUp(n int) {
	reg.Clear(hw.Base+GPIO_PUE, n)
}

// SetIOF configures pin n to use I/O Function override. sel selects between
// IOF0 (sel=0, e.g. UART) and IOF1 (sel=1, e.g. SPI). Enabling IOF
// disconnects normal GPIO input/output logic for that pin.
//
// Example: route UART0 TX (pin 16) and RX (pin 17) through IOF0:
//
//	hw.SetIOF(16, 0)
//	hw.SetIOF(17, 0)
func (hw *GPIO) SetIOF(n int, sel int) {
	// Configure IOF_SEL: 0 = IOF0, 1 = IOF1
	if sel == 0 {
		reg.Clear(hw.iofSELAddr(), n)
	} else {
		reg.Set(hw.iofSELAddr(), n)
	}
	// Enable IOF override for this pin
	reg.Set(hw.iofENAddr(), n)
}

// ClearIOF disables the I/O Function override on pin n, returning it to normal
// GPIO control.
func (hw *GPIO) ClearIOF(n int) {
	reg.Clear(hw.iofENAddr(), n)
}

// SetRiseIRQ enables the rising-edge interrupt on pin n.
func (hw *GPIO) SetRiseIRQ(n int) {
	reg.Set(hw.Base+GPIO_RISE_IE, n)
}

// ClearRiseIRQ disables the rising-edge interrupt on pin n.
func (hw *GPIO) ClearRiseIRQ(n int) {
	reg.Clear(hw.Base+GPIO_RISE_IE, n)
}

// SetFallIRQ enables the falling-edge interrupt on pin n.
func (hw *GPIO) SetFallIRQ(n int) {
	reg.Set(hw.Base+GPIO_FALL_IE, n)
}

// ClearFallIRQ disables the falling-edge interrupt on pin n.
func (hw *GPIO) ClearFallIRQ(n int) {
	reg.Clear(hw.Base+GPIO_FALL_IE, n)
}

// RiseIPending returns true if a rising-edge interrupt is pending on pin n.
func (hw *GPIO) RiseIPending(n int) bool {
	return reg.Read(hw.Base+GPIO_RISE_IP)&pinMask(n) != 0
}

// FallIPending returns true if a falling-edge interrupt is pending on pin n.
func (hw *GPIO) FallIPending(n int) bool {
	return reg.Read(hw.Base+GPIO_FALL_IP)&pinMask(n) != 0
}

// ClearRisePending clears the rising-edge interrupt pending bit for pin n
// (write-1-to-clear).
func (hw *GPIO) ClearRisePending(n int) {
	reg.Write(hw.Base+GPIO_RISE_IP, pinMask(n))
}

// ClearFallPending clears the falling-edge interrupt pending bit for pin n
// (write-1-to-clear).
func (hw *GPIO) ClearFallPending(n int) {
	reg.Write(hw.Base+GPIO_FALL_IP, pinMask(n))
}
