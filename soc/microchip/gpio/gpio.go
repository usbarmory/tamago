// Microchip GPIO support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package gpio implements helpers for GPIO configuration on Microchip SoCs,
// compliant GPIO blocks adopting the following reference specifications:
//   - Microchip - LAN9694/LAN9696/LAN9698 Datasheet - DS00005048E (02-27-25)
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm64` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package gpio

import (
	"errors"
	"fmt"

	"github.com/usbarmory/tamago/internal/reg"
)

// GPIO registers
const (
	GPIO_OUT_SET = 0x00
	GPIO_OUT_CLR = 0x0c
	GPIO_OUT     = 0x18
	GPIO_IN      = 0x24
	GPIO_OE      = 0x30
	GPIO_ALT     = 0x60
)

const (
	bankSize = 32
	banks    = 3
)

// GPIO represents a GPIO controller instance.
type GPIO struct {
	// Base register
	Base uint32
}

// Pin represents a GPIO line instance.
type Pin struct {
	num int
	pos int
	set uint32
	clr uint32
	in  uint32
	oe  uint32
	alt uint32
}

func bank(num int, size uint32) uint32 {
	return uint32(num/bankSize) * size
}

// Init initializes a GPIO line instance.
func (hw *GPIO) Init(num int) (pin *Pin, err error) {
	if hw.Base == 0 {
		return nil, errors.New("invalid GPIO controller instance")
	}

	if num < 0 || num >= banks*bankSize {
		return nil, fmt.Errorf("invalid GPIO number %d", num)
	}

	pin = &Pin{
		num: num,
		pos: num % bankSize,
		set: hw.Base + GPIO_OUT_SET + bank(num, 4),
		clr: hw.Base + GPIO_OUT_CLR + bank(num, 4),
		in:  hw.Base + GPIO_IN + bank(num, 4),
		oe:  hw.Base + GPIO_OE + bank(num, 4),
		alt: hw.Base + GPIO_ALT + bank(num, 4),
	}

	return
}

// Out configures a GPIO line for output.
func (pin *Pin) Out() {
	reg.Set(pin.oe, pin.pos)
}

// In configures a GPIO line for input.
func (pin *Pin) In() {
	reg.Clear(pin.oe, pin.pos)
}

// High configures a GPIO line to be high.
func (pin *Pin) High() {
	reg.Set(pin.set, pin.pos)
}

// Low configures a GPIO line to be low.
func (pin *Pin) Low() {
	reg.Set(pin.clr, pin.pos)
}

// Value returns the GPIO line level.
func (pin *Pin) Value() (high bool) {
	return reg.Get(pin.in, pin.pos)
}

func (pin *Pin) alternateFunctionRegister(bit int) uint32 {
	return pin.alt + uint32(bit*banks*4)
}

// Function selects a GPIO line overlaid function.
func (pin *Pin) Function(mode int) (err error) {
	if mode < 0 || mode > 0b111 {
		return fmt.Errorf("invalid GPIO function %d", mode)
	}

	// Table 3-426: GPIO overlaid functions
	reg.SetTo(pin.alternateFunctionRegister(0), pin.pos, (mode&0b001) > 0)
	reg.SetTo(pin.alternateFunctionRegister(1), pin.pos, (mode&0b010) > 0)
	reg.SetTo(pin.alternateFunctionRegister(2), pin.pos, (mode&0b100) > 0)

	return
}
