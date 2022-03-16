// NXP i.MX6 GPIO driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package imx6

import (
	"fmt"

	"github.com/usbarmory/tamago/internal/reg"
)

// GPIO constants
const (
	GPIO1_BASE = 0x0209c000
	GPIO2_BASE = 0x020a0000
	GPIO3_BASE = 0x020a4000
	GPIO4_BASE = 0x020a8000

	GPIO_DR   = 0x00
	GPIO_GDIR = 0x04

	GPIO_MODE = 5
)

// GPIO instance
type GPIO struct {
	Pad *Pad

	num  int
	data uint32
	dir  uint32
}

// NewGPIO initializes a pad for GPIO mode.
func NewGPIO(num int, instance int, mux uint32, pad uint32) (gpio *GPIO, err error) {
	var base uint32

	if num > 31 {
		return nil, fmt.Errorf("invalid GPIO number %d", num)
	}

	if instance < 1 || instance > 4 {
		return nil, fmt.Errorf("invalid GPIO instance %d", instance)
	}

	switch instance {
	case 1:
		base = GPIO1_BASE
	case 2:
		base = GPIO2_BASE
	case 3:
		base = GPIO3_BASE
	case 4:
		base = GPIO4_BASE
	}

	gpio = &GPIO{
		num:  num,
		data: base + GPIO_DR,
		dir:  base + GPIO_GDIR,
	}

	gpio.Pad, err = NewPad(mux, pad, 0)

	if !Native || err != nil {
		return
	}

	gpio.Pad.Mode(GPIO_MODE)

	return
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

// Value returns the GPIO signal level.
func (gpio *GPIO) Value() (high bool) {
	return reg.Get(gpio.data, gpio.num, 1) == 1
}
