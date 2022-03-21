// BCM2835 SoC GPIO support
// https://github.com/usbarmory/tamago
//
// Copyright (c) the bcm2835 package authors
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package bcm2835

import (
	"fmt"
	"sync"

	"github.com/usbarmory/tamago/arm"
	"github.com/usbarmory/tamago/internal/reg"
)

// GPIO registers
const (
	GPIO_BASE = 0x200000

	GPFSEL0   = GPIO_BASE
	GPFSEL1   = GPIO_BASE + 0x04
	GPSET0    = GPIO_BASE + 0x1c
	GPCLR0    = GPIO_BASE + 0x28
	GPLEV0    = GPIO_BASE + 0x34
	GPPUD     = GPIO_BASE + 0x94
	GPPUDCLK0 = GPIO_BASE + 0x98
)

// GPIO function selections (p92, BCM2835 ARM Peripherals)
const (
	GPIO_INPUT  = 0b000
	GPIO_OUTPUT = 0b001
	GPIO_FN0    = 0b100
	GPIO_FN1    = 0b101
	GPIO_FN2    = 0b110
	GPIO_FN3    = 0b111
	GPIO_FN4    = 0b011
	GPIO_FN5    = 0b010
)

// GPIO instance
type GPIO struct {
	num int
}

type GPIOFunction uint32

var gpmux = sync.Mutex{}

// NewGPIO gets access to a single GPIO line
func NewGPIO(num int) (*GPIO, error) {
	if num > 54 || num < 0 {
		return nil, fmt.Errorf("invalid GPIO number %d", num)
	}

	return &GPIO{num: num}, nil
}

// Out configures a GPIO as output.
func (gpio *GPIO) Out() {
	gpio.SelectFunction(GPIO_OUTPUT)
}

// In configures a GPIO as input.
func (gpio *GPIO) In() {
	gpio.SelectFunction(GPIO_INPUT)
}

// SelectFunction selects the function of a GPIO line.
func (gpio *GPIO) SelectFunction(n GPIOFunction) (err error) {
	if n > 0b111 {
		return fmt.Errorf("invalid GPIO function %d", n)
	}

	register := PeripheralAddress(GPFSEL0 + 4*uint32(gpio.num/10))
	shift := uint32((gpio.num % 10) * 3)
	mask := uint32(0x7 << shift)

	val := reg.Read(register)
	val &= ^(mask)
	val |= (uint32(n) << shift) & mask

	reg.Write(register, val)

	return
}

// GetFunction gets the current function of a GPIO line
func (gpio *GPIO) GetFunction(line int) GPIOFunction {
	val := reg.Read(PeripheralAddress(GPFSEL0 + 4*uint32(gpio.num/10)))
	shift := uint32((gpio.num % 10) * 3)

	return GPIOFunction(val>>shift) & 0x7
}

// High configures a GPIO signal as high.
func (gpio *GPIO) High() {
	register := PeripheralAddress(GPSET0 + 4*uint32(gpio.num/32))
	shift := uint32(gpio.num % 32)

	reg.Write(register, 1<<shift)
}

// Low configures a GPIO signal as low.
func (gpio *GPIO) Low() {
	register := PeripheralAddress(GPCLR0 + 4*uint32(gpio.num/32))
	shift := uint32(gpio.num % 32)

	reg.Write(register, 1<<shift)
}

// Value returns the GPIO signal level.
func (gpio *GPIO) Value() (high bool) {
	register := PeripheralAddress(GPLEV0 + 4*uint32(gpio.num/32))
	shift := uint32(gpio.num % 32)

	return (reg.Read(register)>>shift)&0x1 != 0
}

// PullUpDown controls the pull-up or pull-down state of the line.
//
// The pull-up / pull-down state persists across power-down state
// of the CPU (i.e. always set the pull-up / pull-down to desired
// state before using a GPIO pin).
func (gpio *GPIO) PullUpDown(val uint32) {
	// The control registers are shared between GPIO pins, so
	// hold a mutex for the period.
	gpmux.Lock()
	defer gpmux.Unlock()

	// There is a very specific documented dance (likely related
	// to the persistence over power-down):
	//   1 - write to control register to indicate if wanting to
	//       pull some pins up or down
	//   2 - Wait at least 150 clock cycles for control to setup
	//   3 - Enable the clock for the lines to be modified
	//   4 - Wait at least 150 clock cycles to hold control signal
	//   5 - Remove the control signal
	//   6 - Remove the clock for the line to be modified

	reg.Write(PeripheralAddress(GPPUD), uint32(val))
	arm.Busyloop(150)

	clkRegister := PeripheralAddress(GPPUDCLK0 + 4*uint32(gpio.num/32))
	clkShift := uint32(gpio.num % 32)

	reg.Write(clkRegister, 1<<clkShift)
	arm.Busyloop(150)

	reg.Write(PeripheralAddress(GPPUD), 0)
	reg.Write(clkRegister, 0)
}
