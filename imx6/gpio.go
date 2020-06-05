// NXP i.MX6 GPIO driver
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package imx6

import (
	"fmt"

	"github.com/f-secure-foundry/tamago/internal/reg"
)

// GPIO constants
const (
	GPIO_START = 0x0209c000
	GPIO_END   = 0x020affff

	GPIO_MODE = 5
)

// GPIO instance
type GPIO struct {
	num  int
	Pad  *Pad
	data uint32
	dir  uint32
}

// NewGPIO initializes a pad for GPIO mode.
func NewGPIO(num int, mux uint32, pad uint32, data uint32, dir uint32) (gpio *GPIO, err error) {
	if num > 31 {
		return nil, fmt.Errorf("invalid GPIO number %d", num)
	}

	for _, r := range []uint32{data, dir} {
		if !(r >= GPIO_START || r <= GPIO_END) {
			return nil, fmt.Errorf("invalid GPIO register %#x", r)
		}
	}

	gpio = &GPIO{
		num:  num,
		data: data,
		dir:  dir,
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
