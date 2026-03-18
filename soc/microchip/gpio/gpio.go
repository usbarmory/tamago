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
	"github.com/usbarmory/tamago/internal/reg"
)

// GPIO registers
const (
	GPIO_OUT = 0x18
	GPIO_IN  = 0x24
	GPIO_OE  = 0x30
	GPIO_ALT = 0x60
)

// GPIO controller instance
type GPIO struct {
	// Base register
	Base uint32
}

func addr(num int, size uint32) (off uint32, pos int) {
	switch {
	case num < 32:
		return 0 * size, num
	case num < 64:
		return 1 * size, num % 32
	default:
		return 2 * size, num % 32
	}
}

// Out configures a GPIO as output.
func (gpio *GPIO) Out(num int) {
	off, pos := addr(num, 4)
	reg.Set(gpio.Base+GPIO_OUT+off, pos)
}

// In configures a GPIO as input.
func (gpio *GPIO) In(num int) {
	off, pos := addr(num, 4)
	reg.Clear(gpio.Base+GPIO_OUT+off, pos)
}

// Value returns a GPIO signal level.
func (gpio *GPIO) Value(num int) (high bool) {
	off, pos := addr(num, 4)
	return reg.Get(gpio.Base+GPIO_IN+off, pos)
}

// Function sets a GPIO alternate function assignment.
func (gpio *GPIO) Function(num int, mode int) {
	if mode < 0 || mode > 7 {
		return
	}

	off, pos := addr(num, 12)
	alt := gpio.Base + GPIO_ALT + off

	// Table 3-426: GPIO overlaid functions
	reg.SetTo(alt+(0*4), pos, (mode&0b001) > 0)
	reg.SetTo(alt+(1*4), pos, (mode&0b010) > 0)
	reg.SetTo(alt+(2*4), pos, (mode&0b100) > 0)
}
