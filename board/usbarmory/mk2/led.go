// USB armory Mk II support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package mk2

import (
	"errors"

	"github.com/usbarmory/tamago/soc/imx6/gpio"
	"github.com/usbarmory/tamago/soc/imx6/imx6ul"
	"github.com/usbarmory/tamago/soc/imx6/iomuxc"
)

// LED configuration constants
//
// On the USB armory Mk II the following LEDs are connected:
//   * pad CSI_DATA00, GPIO4_IO21: white
//   * pad CSI_DATA01, GPIO4_IO22: blue
const (
	// GPIO number
	WHITE = 21
	// mux control
	IOMUXC_SW_MUX_CTL_PAD_CSI_DATA00 = 0x020e01e4
	// pad control
	IOMUXC_SW_PAD_CTL_PAD_CSI_DATA00 = 0x020e0470

	// GPIO number
	BLUE = 22
	// mux control
	IOMUXC_SW_MUX_CTL_PAD_CSI_DATA01 = 0x020e01e8
	// pad control
	IOMUXC_SW_PAD_CTL_PAD_CSI_DATA01 = 0x020e0474
)

var white *gpio.Pin
var blue *gpio.Pin

func init() {
	var err error

	ctl := uint32((1 << iomuxc.SW_PAD_CTL_PKE) |
		(iomuxc.SW_PAD_CTL_SPEED_100MHZ << iomuxc.SW_PAD_CTL_SPEED) |
		(iomuxc.SW_PAD_CTL_DSE_2_R0_6 << iomuxc.SW_PAD_CTL_DSE))

	if white, err = imx6ul.GPIO4.InitPad(WHITE,
		IOMUXC_SW_MUX_CTL_PAD_CSI_DATA00, IOMUXC_SW_PAD_CTL_PAD_CSI_DATA00); err != nil {
		panic(err)
	}

	if blue, err = imx6ul.GPIO4.InitPad(BLUE,
		IOMUXC_SW_MUX_CTL_PAD_CSI_DATA01, IOMUXC_SW_PAD_CTL_PAD_CSI_DATA01); err != nil {
		panic(err)
	}

	white.Pad.Ctl(ctl)
	white.Out()

	blue.Pad.Ctl(ctl)
	blue.Out()
}

// LED turns on/off an LED by name.
func LED(name string, on bool) (err error) {
	var led *gpio.Pin

	switch name {
	case "white", "White", "WHITE":
		led = white
	case "blue", "Blue", "BLUE":
		led = blue
	default:
		return errors.New("invalid LED")
	}

	if on {
		led.Low()
	} else {
		led.High()
	}

	return
}
