// Raspberry Pi 2 LED support
// https://github.com/usbarmory/tamago
//
// Copyright (c) the pi2 package authors
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package pi2

import (
	"errors"

	"github.com/usbarmory/tamago/soc/bcm2835"
)

// LED GPIO lines
const (
	// Activity LED
	ACTIVITY = 0x2f
	// Power LED
	POWER = 0x23
)

var activity *bcm2835.GPIO
var power *bcm2835.GPIO

func init() {
	var err error

	activity, err = bcm2835.NewGPIO(ACTIVITY)

	if err != nil {
		panic(err)
	}

	power, err = bcm2835.NewGPIO(POWER)

	if err != nil {
		panic(err)
	}

	activity.Out()
	power.Out()
}

// LED turns on/off an LED by name.
func (b *board) LED(name string, on bool) (err error) {
	var led *bcm2835.GPIO

	switch name {
	case "activity", "Activity", "ACTIVITY":
		led = activity
	case "power", "Power", "POWER":
		led = power
	default:
		return errors.New("invalid LED")
	}

	if on {
		led.High()
	} else {
		led.Low()
	}

	return
}
