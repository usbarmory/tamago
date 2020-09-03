// Raspberry Pi Support
// https://github.com/f-secure-foundry/tamago
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package pi

// Board provides a basic abstraction over the different models of Pi.
type Board interface {
	LEDNames() []string
	LED(name string, on bool) (err error)
}
