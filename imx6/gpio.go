// NXP i.MX6 GPIO driver
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build tamago,arm

package imx6

import (
	"fmt"

	"github.com/f-secure-foundry/tamago/imx6/internal/reg"
)

const (
	IOMUXC_START uint32 = 0x020e0000
	IOMUXC_END   uint32 = 0x0203ffff

	GPIO_START uint32 = 0x0209c000
	GPIO_END   uint32 = 0x020affff

	SW_PAD_CTL_PUS                = 14
	SW_PAD_CTL_PUS_PULL_DOWN_100K = 0
	SW_PAD_CTL_PUS_PULL_UP_47K    = 1
	SW_PAD_CTL_PUS_PULL_UP_100K   = 2
	SW_PAD_CTL_PUS_PULL_UP_22K    = 3

	SW_PAD_CTL_PUE = 13
	SW_PAD_CTL_PKE = 12

	SW_PAD_CTL_SPEED        = 6
	SW_PAD_CTL_SPEED_50MHZ  = 0b00
	SW_PAD_CTL_SPEED_100MHZ = 0b10
	SW_PAD_CTL_SPEED_200MHZ = 0b11

	SW_PAD_CTL_DSE                        = 3
	SW_PAD_CTL_DSE_OUTPUT_DRIVER_DISABLED = 0b000
	SW_PAD_CTL_DSE_2_R0_2                 = 0b010
	SW_PAD_CTL_DSE_2_R0_3                 = 0b011
	SW_PAD_CTL_DSE_2_R0_4                 = 0b100
	SW_PAD_CTL_DSE_2_R0_5                 = 0b101
	SW_PAD_CTL_DSE_2_R0_6                 = 0b110
	SW_PAD_CTL_DSE_2_R0_7                 = 0b111

	MUX_MODE = 0
	ALT5     = 0b0101
)

type GPIO struct {
	num  int
	mux  uint32
	pad  uint32
	data uint32
	dir  uint32
}

// NewGPIO initializes a pad for GPIO mode.
func NewGPIO(num int, mux uint32, pad uint32, data uint32, dir uint32) (gpio *GPIO, err error) {
	if num > 31 {
		return nil, fmt.Errorf("invalid GPIO number %d", num)
	}

	for _, r := range []uint32{mux, pad} {
		if !(r >= IOMUXC_START || r <= IOMUXC_END) {
			return nil, fmt.Errorf("invalid IOMUXC register %x", r)
		}
	}

	for _, r := range []uint32{data, dir} {
		if !(r >= GPIO_START || r <= GPIO_END) {
			return nil, fmt.Errorf("invalid GPIO register %x", r)
		}
	}

	gpio = &GPIO{
		num:  num,
		mux:  mux,
		pad:  pad,
		data: data,
		dir:  dir,
	}

	if !Native {
		return
	}

	// select GPIO mode
	reg.SetN(gpio.mux, MUX_MODE, 0b1111, ALT5)

	return
}

// PadCtl configures the GPIO pad by setting the desired value to the
// SW_PAD_CTL register.
func (gpio *GPIO) PadCtl(val uint32) {
	reg.Write(gpio.pad, val)
}

// Out configures a GPIO as output.
func (gpio *GPIO) Out() {
	reg.Set(gpio.dir, gpio.num)
}

// In configures a GPIO as input.
func (gpio *GPIO) In() {
	reg.Clear(gpio.dir, gpio.num)
}

// High configures a GPIO signal as high.
func (gpio *GPIO) High() {
	reg.Set(gpio.data, gpio.num)
}

// Low configures a GPIO signal as low.
func (gpio *GPIO) Low() {
	reg.Clear(gpio.data, gpio.num)
}
