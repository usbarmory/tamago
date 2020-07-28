// Raspberry Pi GPIO Support
// https://github.com/f-secure-foundry/tamago
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package pi

import (
	"fmt"

	"github.com/f-secure-foundry/tamago/bcm2835"
	"github.com/f-secure-foundry/tamago/internal/reg"
)

const gpfsel0 uint32 = 0x200000
const gpset0 uint32 = 0x20001C
const gpclr0 uint32 = 0x200028

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

type gpio struct {
}

// GPIO provides convenient access to the Raspberry Pi GPIO lines
var GPIO = gpio{}

// SelectFunction selects the function of a GPIO line
func (p gpio) SelectFunction(line int, fn GPIOFunction) error {
	if line > 54 || line < 0 || fn > GPIOFunctionAltFunction5 {
		return fmt.Errorf("invalid parameter")
	}

	register := bcm2835.PeripheralBase + gpfsel0 + 4*uint32(line/10)
	shift := uint32((line % 10) * 3)
	mask := uint32(0x7 << shift)

	val := reg.Read(register)
	val &= ^(mask)
	val |= (uint32(fn) << shift) & mask
	reg.Write(register, val)

	return nil
}

// GetFunction gets the current function of a GPIO line
func (p gpio) GetFunction(line int) (GPIOFunction, error) {
	if line > 54 || line < 0 {
		return GPIOFunctionInput, fmt.Errorf("invalid parameter")
	}

	register := bcm2835.PeripheralBase + gpfsel0 + 4*uint32(line/10)
	shift := uint32((line % 10) * 3)
	val := reg.Read(register)
	return GPIOFunction((val >> shift) & 0x7), nil
}

// Set the status of a GPIO line to 'on'
func (p gpio) Set(line int) error {
	if line > 54 || line < 0 {
		return fmt.Errorf("invalid parameter")
	}

	register := bcm2835.PeripheralBase + gpset0 + 4*uint32(line/32)
	shift := uint32(line % 32)

	reg.Write(register, 1<<shift)

	return nil
}

// Set the status of a GPIO line to 'off'
func (p gpio) Clear(line int) error {
	if line > 54 || line < 0 {
		return fmt.Errorf("invalid parameter")
	}

	register := bcm2835.PeripheralBase + gpclr0 + 4*uint32(line/32)
	shift := uint32(line % 32)
	reg.Write(register, 1<<shift)

	return nil
}
