// Raspberry Pi LED Support
// https://github.com/f-secure-foundry/tamago
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package pi

// LEDType is the purpose of the LED
type LEDType int

const (
	// ActivityLED is the green LED
	ActivityLED LEDType = 0

	// PowerLED is the red LED found on some boards
	PowerLED LEDType = 1
)

// LED represents a single LED on the board
type LED struct {
	Type LEDType
	Line int
}

// Init configures the GPIO to control the LED
func (led *LED) Init() {
	GPIO.SelectFunction(led.Line, GPIOFunctionOutput)
}

// On turns the LED on
func (led *LED) On() {
	led.SetState(true)
}

// Off turns the LED off
func (led *LED) Off() {
	led.SetState(false)
}

// SetState sets the LED on or off
func (led *LED) SetState(state bool) {
	if state {
		GPIO.Set(led.Line)
	} else {
		GPIO.Clear(led.Line)
	}
}
