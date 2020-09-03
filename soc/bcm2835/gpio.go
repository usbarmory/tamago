// BCM2835 SOC GPIO Support
// https://github.com/f-secure-foundry/tamago
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package bcm2835

import (
	"fmt"

	"github.com/f-secure-foundry/tamago/internal/reg"
)

const (
	gpfsel0 = 0x200000
	gpset0  = 0x20001C
	gpclr0  = 0x200028
	gplev0  = 0x200034
)

// GPIOFunction represents the modes of a GPIO line
type GPIOFunction uint32

const (
	// GPIOFunctionInput uses the GPIO line for input
	GPIOFunctionInput GPIOFunction = 0

	// GPIOFunctionOutput uses the GPIO line for output
	GPIOFunctionOutput = 1

	// GPIOFunctionAltFunction0 for it alternate function 0
	GPIOFunctionAltFunction0 = 2

	// GPIOFunctionAltFunction1 for it alternate function 1
	GPIOFunctionAltFunction1 = 3

	// GPIOFunctionAltFunction2 for it alternate function 2
	GPIOFunctionAltFunction2 = 4

	// GPIOFunctionAltFunction3 for it alternate function 3
	GPIOFunctionAltFunction3 = 5

	// GPIOFunctionAltFunction4 for it alternate function 4
	GPIOFunctionAltFunction4 = 6

	// GPIOFunctionAltFunction5 for it alternate function 5
	GPIOFunctionAltFunction5 = 7
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
