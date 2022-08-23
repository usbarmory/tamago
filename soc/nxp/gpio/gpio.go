// NXP GPIO support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package gpio implements helpers for GPIO configuration on NXP SoCs.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/usbarmory/tamago.
package gpio

import (
	"errors"
	"fmt"

	"github.com/usbarmory/tamago/internal/reg"
)

// GPIO registers
const (
	GPIO_DR   = 0x00
	GPIO_GDIR = 0x04
)

// GPIO controller instance
type GPIO struct {
	// Controller index
	Index int
	// Base register
	Base uint32
}

// Pin instance
type Pin struct {
	num  int
	data uint32
	dir  uint32
}

// Init initializes a GPIO.
func (hw *GPIO) Init(num int) (gpio *Pin, err error) {
	if hw.Base == 0 {
		return nil, errors.New("invalid GPIO controller instance")
	}

	if num > 31 {
		return nil, fmt.Errorf("invalid GPIO number %d", num)
	}

	gpio = &Pin{
		num:  num,
		data: hw.Base + GPIO_DR,
		dir:  hw.Base + GPIO_GDIR,
	}

	return
}

// Out configures a GPIO as output.
func (gpio *Pin) Out() {
	reg.Set(gpio.dir, gpio.num)
}

// In configures a GPIO as input.
func (gpio *Pin) In() {
	reg.Clear(gpio.dir, gpio.num)
}

// High configures a GPIO signal as high.
func (gpio *Pin) High() {
	reg.Set(gpio.data, gpio.num)
}

// Low configures a GPIO signal as low.
func (gpio *Pin) Low() {
	reg.Clear(gpio.data, gpio.num)
}

// Value returns the GPIO signal level.
func (gpio *Pin) Value() (high bool) {
	return reg.Get(gpio.data, gpio.num, 1) == 1
}
