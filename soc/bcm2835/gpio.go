// BCM2835 SoC GPIO Support
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) the bcm2835 package authors
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package bcm2835

import (
	"fmt"
	"sync"

	"github.com/f-secure-foundry/tamago/arm"
	"github.com/f-secure-foundry/tamago/internal/reg"
)

const (
	gpfsel0   = 0x200000
	gpset0    = 0x20001C
	gpclr0    = 0x200028
	gplev0    = 0x200034
	gppud     = 0x200094
	gppudclk0 = 0x200098
)

var pullUpDownControlLock = sync.Mutex{}

// GPIOFunction represents the modes of a GPIO line
type GPIOFunction uint32

const (
	// GPIOFunctionInput uses the GPIO line for input
	GPIOFunctionInput GPIOFunction = 0x0

	// GPIOFunctionOutput uses the GPIO line for output
	GPIOFunctionOutput = 0x1

	// GPIOFunctionAltFunction0 for it alternate function 0
	GPIOFunctionAltFunction0 = 0x4

	// GPIOFunctionAltFunction1 for it alternate function 1
	GPIOFunctionAltFunction1 = 0x5

	// GPIOFunctionAltFunction2 for it alternate function 2
	GPIOFunctionAltFunction2 = 0x6

	// GPIOFunctionAltFunction3 for it alternate function 3
	GPIOFunctionAltFunction3 = 0x7

	// GPIOFunctionAltFunction4 for it alternate function 4
	GPIOFunctionAltFunction4 = 0x3

	// GPIOFunctionAltFunction5 for it alternate function 5
	GPIOFunctionAltFunction5 = 0x2
)

// GPIOPullUpDown controls the pull-up or pull-down resistor
// state of the GPIO line
type GPIOPullUpDown uint32

const (
	// GPIONoPullupOrPullDown disables any pull-up or pull-down
	GPIONoPullupOrPullDown GPIOPullUpDown = 0

	// GPIOPullDown applies a pull-down resistor on the GPIO line
	GPIOPullDown = 1

	// GPIOPullUp applies a pull-up resistor on the GPIO line
	GPIOPullUp = 2
)

// GPIO instance
type GPIO struct {
	num int
}

// NewGPIO gets access to a single GPIO line
func NewGPIO(num int) (*GPIO, error) {
	if num > 54 || num < 0 {
		return nil, fmt.Errorf("invalid GPIO number %d", num)
	}

	return &GPIO{num: num}, nil
}

// Out configures a GPIO as output.
func (gpio *GPIO) Out() {
	gpio.SelectFunction(GPIOFunctionOutput)
}

// In configures a GPIO as input.
func (gpio *GPIO) In() {
	gpio.SelectFunction(GPIOFunctionInput)
}

// SelectFunction selects the function of a GPIO line
func (gpio *GPIO) SelectFunction(fn GPIOFunction) error {
	if fn > GPIOFunctionAltFunction5 {
		return fmt.Errorf("invalid GPIO function %d", fn)
	}

	register := PeripheralAddress(gpfsel0 + 4*uint32(gpio.num/10))
	shift := uint32((gpio.num % 10) * 3)
	mask := uint32(0x7 << shift)

	val := reg.Read(register)
	val &= ^(mask)
	val |= (uint32(fn) << shift) & mask
	reg.Write(register, val)

	return nil
}

// GetFunction gets the current function of a GPIO line
func (gpio *GPIO) GetFunction(line int) (GPIOFunction, error) {
	register := PeripheralAddress(gpfsel0 + 4*uint32(gpio.num/10))
	shift := uint32((gpio.num % 10) * 3)
	val := reg.Read(register)
	return GPIOFunction((val >> shift) & 0x7), nil
}

// High configures a GPIO signal as high.
func (gpio *GPIO) High() {
	register := PeripheralAddress(gpset0 + 4*uint32(gpio.num/32))
	shift := uint32(gpio.num % 32)
	reg.Write(register, 1<<shift)
}

// Low configures a GPIO signal as low.
func (gpio *GPIO) Low() {
	register := PeripheralAddress(gpclr0 + 4*uint32(gpio.num/32))
	shift := uint32(gpio.num % 32)
	reg.Write(register, 1<<shift)
}

// Value returns the GPIO signal level.
func (gpio *GPIO) Value() (high bool) {
	register := PeripheralAddress(gplev0 + 4*uint32(gpio.num/32))
	shift := uint32(gpio.num % 32)
	return (reg.Read(register)>>shift)&0x1 != 0
}

// PullUpDown controls the pull-up or pull-down state of the line.
//
// The pull-up / pull-down state persists across power-down state
// of the CPU (i.e. always set the pull-up / pull-down to desired
// state before using a GPIO pin)
func (gpio *GPIO) PullUpDown(value GPIOPullUpDown) {
	// There is a very specific documented dance (likely related
	// to the persistence over power-down):
	//   1 - write to control register to indicate if wanting to
	//       pull some pins up or down
	//   2 - Wait at least 150 clock cycles for control to setup
	//   3 - Enable the clock for the lines to be modified
	//   4 - Wait at least 150 clock cycles to hold control signal
	//   5 - Remove the control signal
	//   6 - Remove the clock for the line to be modified

	// The control registers are shared between GPIO pins, so
	// hold a mutex for the period.
	pullUpDownControlLock.Lock()
	defer pullUpDownControlLock.Unlock()

	reg.Write(PeripheralAddress(gppud), uint32(value))

	arm.Busyloop(150)

	clkRegister := PeripheralAddress(gppudclk0 + 4*uint32(gpio.num/32))
	clkShift := uint32(gpio.num % 32)
	reg.Write(clkRegister, 1<<clkShift)

	arm.Busyloop(150)

	reg.Write(gppud, 0)
	reg.Write(clkRegister, 0)
}
