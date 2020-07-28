// Raspberry Pi Support
// https://github.com/f-secure-foundry/tamago
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package pi

// Board provides a basic abstraction over the different models of Pi.
type Board interface {
	// LEDs gets the LEDs present on the board
	//
	// The board LEDs available varies by model, the Pi Zero has a single
	// 'activity' LED (no 'power' LED), the Pi 2 has both an 'activity'
	// and a 'power' LED.
	LEDs() []LED
}
