// Raspberry Pi 2 Support
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) the pi2 package authors
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package pi2

import (
	"errors"

	"github.com/f-secure-foundry/tamago/board/pi-foundation"
	"github.com/f-secure-foundry/tamago/soc/bcm2835"
)

const (
	gpioLineActivityLED = 0x2f
	gpioLinePowerLED    = 0x23
)

type board struct{}

// Board provides access to the capabilities of the Pi2.
var Board pi.Board = &board{}

var activityLED *bcm2835.GPIO
var powerLED *bcm2835.GPIO

func init() {
	var err error
	activityLED, err = bcm2835.NewGPIO(gpioLineActivityLED)
	if err != nil {
		panic(err)
	}

	powerLED, err = bcm2835.NewGPIO(gpioLinePowerLED)
	if err != nil {
		panic(err)
	}
}

func (b *board) LEDNames() []string {
	return []string{"activity", "power"}
}

func (b *board) LED(name string, on bool) (err error) {
	var led *bcm2835.GPIO

	switch name {
	case "activity", "Activity", "ACTIVITY":
		led = activityLED
	case "power", "Power", "POWER":
		led = powerLED
	default:
		return errors.New("invalid LED")
	}

	led.Out()
	if on {
		led.High()
	} else {
		led.Low()
	}

	return nil
}
