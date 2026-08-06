// USB armory Mk II support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package mk2

import (
	"errors"
	"strings"

	"github.com/usbarmory/tamago/soc/nxp/gpio"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
	"github.com/usbarmory/tamago/soc/nxp/iomuxc"
)

// LED configuration constants
//
// On the USB armory Mk II the following LEDs are connected:
//   - pad CSI_DATA00, GPIO4_IO21: white
//   - pad CSI_DATA01, GPIO4_IO22: blue
//
// On the USB armory Mk II LAN the RJ45 connector LEDs can be controlled
// through the Ethernet PHY.
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

var (
	white *gpio.Pin
	blue  *gpio.Pin
)

func init() {
	var err error

	ctl := uint32((1 << iomuxc.SW_PAD_CTL_PKE) |
		(iomuxc.SW_PAD_CTL_SPEED_100MHZ << iomuxc.SW_PAD_CTL_SPEED) |
		(iomuxc.SW_PAD_CTL_DSE_2_R0_6 << iomuxc.SW_PAD_CTL_DSE))

	if white, err = imx6ul.GPIO4.Init(WHITE); err != nil {
		panic(err)
	}

	white.Out()

	p := iomuxc.Init(
		IOMUXC_SW_MUX_CTL_PAD_CSI_DATA00,
		IOMUXC_SW_PAD_CTL_PAD_CSI_DATA00,
		GPIO_MODE)
	p.Ctl(ctl)

	if blue, err = imx6ul.GPIO4.Init(BLUE); err != nil {
		panic(err)
	}

	blue.Out()

	p = iomuxc.Init(
		IOMUXC_SW_MUX_CTL_PAD_CSI_DATA01,
		IOMUXC_SW_PAD_CTL_PAD_CSI_DATA01,
		GPIO_MODE)
	p.Ctl(ctl)
}

// LED turns on/off an LED by name.
func LED(name string, on bool) (err error) {
	var led *gpio.Pin

	switch {
	case strings.EqualFold(name, "white"):
		led = white
	case strings.EqualFold(name, "blue"):
		led = blue
	case strings.EqualFold(name, "green") && phy != nil:
		return phy.LED(1, on)
	case strings.EqualFold(name, "yellow") && phy != nil:
		return phy.LED(2, on)
	default:
		return errors.New("invalid LED")
	}

	if led != nil {
		if on {
			led.Low()
		} else {
			led.High()
		}
	}

	return
}
