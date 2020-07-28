// Raspberry Pi Zero Support
// https://github.com/f-secure-foundry/tamago
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package pizero

import "github.com/f-secure-foundry/tamago/pi"

type board struct{}

// Board provides access to the capabilities of the Pi Zero.
var Board pi.Board = &board{}

func (b *board) LEDs() []pi.LED {
	return []pi.LED{{
		Type: pi.ActivityLED,
		Line: 0x2f,
	}}
}
