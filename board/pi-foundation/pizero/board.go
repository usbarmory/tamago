// Raspberry Pi Zero Support
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) the pizero package authors
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package pizero

import (
	"errors"

	"github.com/f-secure-foundry/tamago/board/pi-foundation"
	"github.com/f-secure-foundry/tamago/soc/bcm2835"
)

const (
	gpioLineActivityLED = 0x2f
)

type board struct{}

// Board provides access to the capabilities of the Pi Zero.
var Board pi.Board = &board{}

var activityLED *bcm2835.GPIO

func init() {
	var err error
	activityLED, err = bcm2835.NewGPIO(gpioLineActivityLED)
	if err != nil {
		panic(err)
	}
}

func (b *board) LEDNames() []string {
	return []string{"activity"}
}

func (b *board) LED(name string, on bool) (err error) {
	var led *bcm2835.GPIO

	switch name {
	case "activity", "Activity", "ACTIVITY":
		led = activityLED
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
