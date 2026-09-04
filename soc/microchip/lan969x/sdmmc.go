// Microchip LAN969x configuration and support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package lan969x

import "github.com/usbarmory/tamago/soc/microchip/gpio"

const (
	gckConfigOffset = 0xb4
	sdmmc0ClockID   = 2

	sdmmc0ParentClock = 1_000_000_000
	sdmmc0TargetClock = 200_000_000

	sdmmc0GPIOStart    = 14
	sdmmc0GPIOEnd      = 24
	sdmmc0GPIOFunction = 1
)

func configureSDMMC0Pins() (err error) {
	var pin *gpio.Pin

	for num := sdmmc0GPIOStart; num <= sdmmc0GPIOEnd; num++ {
		if pin, err = GPIO.Init(num); err != nil {
			return
		}
		if err = pin.Function(sdmmc0GPIOFunction); err != nil {
			return
		}
	}

	return
}
