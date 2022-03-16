// Raspberry Pi Zero LED support
// https://github.com/usbarmory/tamago
//
// Copyright (c) the pizero package authors
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package pizero

import (
	"errors"

	"github.com/usbarmory/tamago/soc/bcm2835"
)

// LED GPIO lines
const (
	// Activity LED
	ACTIVITY = 0x2f
)

var activity *bcm2835.GPIO

func init() {
	var err error

	activity, err = bcm2835.NewGPIO(ACTIVITY)

	if err != nil {
		panic(err)
	}

	activity.Out()
}

// LED turns on/off an LED by name.
func (b *board) LED(name string, on bool) (err error) {
	var led *bcm2835.GPIO

	switch name {
	case "activity", "Activity", "ACTIVITY":
		led = activity
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
