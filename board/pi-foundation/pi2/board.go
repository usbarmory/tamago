// Raspberry Pi 2 Support
// https://github.com/f-secure-foundry/tamago
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package pi2

import "github.com/f-secure-foundry/tamago/board/pi-foundation"

const (
	gpioLineActivityLED = 0x2f
	gpioLinePowerLED    = 0x23
)

type board struct{}

// Board provides access to the capabilities of the Pi2.
var Board pi.Board = &board{}

func (b *board) LEDs() []pi.LED {
	return []pi.LED{
		{
			Type: pi.ActivityLED,
			Line: gpioLineActivityLED,
		},
		{
			Type: pi.PowerLED,
			Line: gpioLinePowerLED,
		},
	}
}
